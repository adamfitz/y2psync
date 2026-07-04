package sync

import "github.com/adam/y2psync/internal/database"

type MergeEngine struct {
	db *database.DB
}

func NewMergeEngine(db *database.DB) *MergeEngine {
	return &MergeEngine{db: db}
}

func (m *MergeEngine) ResolveConflicts(msg *SyncMessage) error {
	return MergeIncomingData(m.db, msg)
}
