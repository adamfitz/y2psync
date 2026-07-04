package sync

import "fmt"

type Session struct{}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) SyncWithPeer(peerAddr string) error {
	return fmt.Errorf("sync session not implemented: requires Noise protocol handshake and protobuf serialization")
}
