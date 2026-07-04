package sync

import (
	"context"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/adam/y2psync/internal/database"
)

type Session struct {
	host host.Host
	db   *database.DB
}

func NewSession(h host.Host, db *database.DB) *Session {
	return &Session{host: h, db: db}
}

func (s *Session) SyncWithPeer(ctx context.Context, pi peer.AddrInfo, localPeers []string) ([]string, error) {
	if err := s.host.Connect(ctx, pi); err != nil {
		return nil, fmt.Errorf("connect to peer: %w", err)
	}

	stream, err := s.host.NewStream(ctx, pi.ID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}
	defer stream.Close()

	return s.exchange(ctx, stream, localPeers)
}

func (s *Session) exchange(ctx context.Context, stream network.Stream, localPeers []string) ([]string, error) {
	localData, err := collectLocalData(
		database.NewPlaylistRepo(s.db),
		database.NewSubscriptionRepo(s.db),
		s.host.ID().String(),
		localPeers,
	)
	if err != nil {
		return nil, fmt.Errorf("collect local data: %w", err)
	}

	if err := encodeJSONMessage(stream, localData); err != nil {
		return nil, fmt.Errorf("send local data: %w", err)
	}

	remoteMsg, err := decodeJSONMessage(stream)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("receive remote data: %w", err)
	}

	if err := MergeIncomingData(s.db, remoteMsg); err != nil {
		return nil, fmt.Errorf("merge remote data: %w", err)
	}

	return remoteMsg.KnownPeers, nil
}
