package model

import "time"

type SyncMetadata struct {
	PeerID            string    `json:"peer_id"`
	LastSyncTimestamp time.Time `json:"last_sync_timestamp"`
	LastSyncStatus    string    `json:"last_sync_status"`
}

type PeerConfig struct {
	PeerID        string `json:"peer_id"`
	MasterSyncKey string `json:"-"` // never persisted in plaintext
	SyncGroupKey  []byte `json:"-"`
	RendezvousTag string `json:"-"`
}
