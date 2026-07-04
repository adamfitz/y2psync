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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/adam/y2psync/internal/database"
)

type Status string

const (
	StatusIdle     Status = "Idle"
	StatusStarting Status = "Starting..."
	StatusRunning  Status = "Running"
	StatusStopped  Status = "Stopped"
)

type Syncer struct {
	db         *database.DB
	configRepo *database.ConfigRepo
	rendezvous string
	peerIDStr  string

	mu        sync.Mutex
	status    Status
	host      host.Host
	dht       *dht.IpfsDHT
	discovery *Discovery
	session   *Session
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	statusChan chan Status
	peerCount  int
}

func NewSyncer(db *database.DB, configRepo *database.ConfigRepo) *Syncer {
	peerID, _ := configRepo.Get("peer_id")
	rendezvous, _ := configRepo.Get("rendezvous_tag")

	return &Syncer{
		db:         db,
		configRepo: configRepo,
		peerIDStr:  peerID,
		rendezvous: rendezvous,
		status:     StatusIdle,
		statusChan: make(chan Status, 8),
	}
}

func (s *Syncer) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *Syncer) PeerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.peerCount
}

func (s *Syncer) StatusChan() <-chan Status {
	return s.statusChan
}

func (s *Syncer) setStatus(st Status) {
	s.mu.Lock()
	s.status = st
	s.mu.Unlock()
	select {
	case s.statusChan <- st:
	default:
	}
}

func (s *Syncer) IsSyncConfigured() bool {
	configured, _ := s.configRepo.Get("sync_key_configured")
	return configured == "true"
}

func (s *Syncer) Start() error {
	s.mu.Lock()
	if s.status == StatusRunning || s.status == StatusStarting {
		s.mu.Unlock()
		return fmt.Errorf("sync already running")
	}
	s.status = StatusStarting
	s.mu.Unlock()

	if !s.IsSyncConfigured() {
		return fmt.Errorf("sync key not configured")
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	privKey, err := s.deriveKey()
	if err != nil {
		cancel()
		return fmt.Errorf("derive key: %w", err)
	}

	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")

	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(privKey),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(),
	)
	if err != nil {
		cancel()
		return fmt.Errorf("create libp2p host: %w", err)
	}

	d, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		h.Close()
		cancel()
		return fmt.Errorf("create dht: %w", err)
	}

	if err := d.Bootstrap(ctx); err != nil {
		h.Close()
		cancel()
		return fmt.Errorf("bootstrap dht: %w", err)
	}

	sess := NewSession(h, s.db)
	h.SetStreamHandler(ProtocolID, sess.handleStream)

	disc := NewDiscovery(h, d, s.rendezvous)
	if err := disc.Start(ctx); err != nil {
		h.Close()
		cancel()
		return fmt.Errorf("start discovery: %w", err)
	}

	s.mu.Lock()
	s.host = h
	s.dht = d
	s.discovery = disc
	s.session = sess
	s.mu.Unlock()

	s.wg.Add(1)
	go s.syncLoop(ctx, disc)

	s.setStatus(StatusRunning)

	return nil
}

func (s *Syncer) Stop() {
	s.mu.Lock()
	wasRunning := s.status == StatusRunning || s.status == StatusStarting
	s.mu.Unlock()

	if !wasRunning {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()

	s.mu.Lock()
	if s.discovery != nil {
		s.discovery.Stop()
	}
	if s.host != nil {
		s.host.Close()
	}
	s.discovery = nil
	s.host = nil
	s.dht = nil
	s.session = nil
	s.peerCount = 0
	s.mu.Unlock()

	s.setStatus(StatusStopped)
	time.Sleep(100 * time.Millisecond)
	s.setStatus(StatusIdle)
}

func (s *Syncer) syncLoop(ctx context.Context, disc *Discovery) {
	defer s.wg.Done()

	peerChan := disc.Peers()
	recent := make(map[peer.ID]time.Time)

	for {
		select {
		case <-ctx.Done():
			return
		case pi, ok := <-peerChan:
			if !ok {
				return
			}
			if time.Since(recent[pi.ID]) < time.Minute {
				continue
			}
			recent[pi.ID] = time.Now()

			func() {
				sess := NewSession(s.host, s.db)
				if err := sess.SyncWithPeer(ctx, pi); err != nil {
					return
				}

				s.mu.Lock()
				s.peerCount = len(recent)
				s.mu.Unlock()
			}()
		}
	}
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
