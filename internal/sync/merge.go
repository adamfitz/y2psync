package sync

import "fmt"

type MergeEngine struct{}

func NewMergeEngine() *MergeEngine {
	return &MergeEngine{}
}

func (m *MergeEngine) ResolveConflicts() error {
	return fmt.Errorf("merge engine not implemented: requires full sync protocol")
}
