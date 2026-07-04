package sync

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"

	"github.com/adam/y2psync/internal/database"
)

type Status string

const (
	StatusIdle        Status = "Idle"
	StatusDiscovering Status = "Discovering..."
	StatusSettled     Status = "Settled"
	StatusError       Status = "Error"
)

const (
	activeInterval  = 30 * time.Second
	settledInterval = 5 * time.Minute
	settleCycles    = 3
	reSyncInterval  = 5 * time.Minute
)

type SyncStatus struct {
	State       Status
	KnownPeers  int
	SyncedPeers int
	LastSync    string
}

type Syncer struct {
	db         *database.DB
	configRepo *database.ConfigRepo
	rendezvous string
	peerIDStr  string

	mu          sync.Mutex
	status      Status
	host        host.Host
	dht         *dht.IpfsDHT
	discovery   *Discovery
	knownPeers  map[string]bool
	syncedPeers map[string]time.Time
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool

	statusChan chan SyncStatus
}

func NewSyncer(db *database.DB, configRepo *database.ConfigRepo) *Syncer {
	peerID, _ := configRepo.Get("peer_id")
	rendezvous, _ := configRepo.Get("rendezvous_tag")

	return &Syncer{
		db:          db,
		configRepo:  configRepo,
		peerIDStr:   peerID,
		rendezvous:  rendezvous,
		status:      StatusIdle,
		knownPeers:  make(map[string]bool),
		syncedPeers: make(map[string]time.Time),
		statusChan:  make(chan SyncStatus, 8),
	}
}

func (s *Syncer) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *Syncer) SyncStatus() SyncStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	lastSync, _ := s.configRepo.Get("last_sync_timestamp")
	return SyncStatus{
		State:       s.status,
		KnownPeers:  len(s.knownPeers),
		SyncedPeers: len(s.syncedPeers),
		LastSync:    lastSync,
	}
}

func (s *Syncer) SyncStatusChan() <-chan SyncStatus {
	return s.statusChan
}

func (s *Syncer) setStatus(st Status) {
	s.mu.Lock()
	s.status = st
	s.mu.Unlock()
	s.sendStatus()
}

func (s *Syncer) sendStatus() {
	lastSync, _ := s.configRepo.Get("last_sync_timestamp")
	s.mu.Lock()
	ss := SyncStatus{
		State:       s.status,
		KnownPeers:  len(s.knownPeers),
		SyncedPeers: len(s.syncedPeers),
		LastSync:    lastSync,
	}
	s.mu.Unlock()
	select {
	case s.statusChan <- ss:
	default:
	}
}

func (s *Syncer) IsSyncConfigured() bool {
	configured, _ := s.configRepo.Get("sync_key_configured")
	return configured == "true"
}

func (s *Syncer) Run() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	if !s.IsSyncConfigured() {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	privKey, err := s.deriveKey()
	if err != nil {
		s.setStatus(StatusError)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}

	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(),
	)
	if err != nil {
		s.setStatus(StatusError)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	d, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		h.Close()
		cancel()
		s.setStatus(StatusError)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}

	if err := d.Bootstrap(ctx); err != nil {
		h.Close()
		cancel()
		s.setStatus(StatusError)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}

	disc := NewDiscovery(h, d, s.rendezvous)

	h.SetStreamHandler(ProtocolID, func(stream network.Stream) {
		defer stream.Close()
		remotePeers, err := NewSession(h, s.db).exchange(context.Background(), stream, s.getKnownPeers())
		if err != nil {
			return
		}
		remoteID := stream.Conn().RemotePeer().String()
		s.mu.Lock()
		s.knownPeers[remoteID] = true
		s.syncedPeers[remoteID] = time.Now()
		for _, p := range remotePeers {
			if p != s.peerIDStr {
				s.knownPeers[p] = true
			}
		}
		s.configRepo.Set("last_sync_timestamp", time.Now().UTC().Format(time.RFC3339))
		s.mu.Unlock()
		s.sendStatus()
	})

	s.mu.Lock()
	s.host = h
	s.dht = d
	s.discovery = disc
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go s.discoveryLoop(ctx, disc)

	s.setStatus(StatusDiscovering)
}

func (s *Syncer) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()

	s.mu.Lock()
	if s.host != nil {
		s.host.Close()
	}
	s.host = nil
	s.dht = nil
	s.discovery = nil
	s.cancel = nil
	s.knownPeers = make(map[string]bool)
	s.syncedPeers = make(map[string]time.Time)
	s.running = false
	s.mu.Unlock()

	s.setStatus(StatusIdle)
}

func (s *Syncer) discoveryLoop(ctx context.Context, disc *Discovery) {
	defer s.wg.Done()

	interval := activeInterval
	noNewCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}

		disc.Advertise(ctx)

		peers, err := disc.FindPeers(ctx)
		if err != nil {
			continue
		}

		hostID := s.host.ID().String()
		foundNew := false

		for _, pi := range peers {
			pid := pi.ID.String()
			if pid == hostID {
				continue
			}

			s.mu.Lock()
			alreadyKnown := s.knownPeers[pid]
			if !alreadyKnown {
				s.knownPeers[pid] = true
				foundNew = true
			}
			lastSync := s.syncedPeers[pid]
			needsResync := time.Since(lastSync) > reSyncInterval
			s.mu.Unlock()

			if !needsResync && alreadyKnown {
				continue
			}

			func() {
				sess := NewSession(s.host, s.db)
				remotePeers, err := sess.SyncWithPeer(ctx, pi, s.getKnownPeers())
				if err != nil {
					return
				}

				now := time.Now()
				s.mu.Lock()
				s.syncedPeers[pid] = now
				for _, p := range remotePeers {
					if p != s.peerIDStr {
						s.knownPeers[p] = true
					}
				}
				s.configRepo.Set("last_sync_timestamp", now.UTC().Format(time.RFC3339))
				s.mu.Unlock()
				s.sendStatus()
			}()
		}

		if foundNew {
			noNewCount = 0
			interval = activeInterval
			s.setStatus(StatusDiscovering)
		} else {
			noNewCount++
			if noNewCount >= settleCycles {
				interval = settledInterval
				s.setStatus(StatusSettled)
			}
		}
		s.sendStatus()
	}
}

func (s *Syncer) getKnownPeers() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	peers := make([]string, 0, len(s.knownPeers))
	for k := range s.knownPeers {
		peers = append(peers, k)
	}
	return peers
}

func (s *Syncer) deriveKey() (crypto.PrivKey, error) {
	seedHex, err := s.configRepo.Get("peer_id")
	if err != nil || seedHex == "" {
		return nil, fmt.Errorf("no peer id configured")
	}

	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("decode peer id: %w", err)
	}

	keyDigest := sha256.Sum256(append([]byte("libp2p-ed25519"), seed...))
	privKey := ed25519.NewKeyFromSeed(keyDigest[:])

	libp2pKey, _, err := crypto.KeyPairFromStdKey(&privKey)
	if err != nil {
		return nil, fmt.Errorf("convert to libp2p key: %w", err)
	}

	return libp2pKey, nil
}
