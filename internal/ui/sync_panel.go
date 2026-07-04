package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/adam/y2psync/internal/database"
)

type SyncPanel struct {
	configRepo  *database.ConfigRepo
	statusLabel *widget.Label
	peerLabel   *widget.Label
	lastSync    *widget.Label
}

func NewSyncPanel(configRepo *database.ConfigRepo) *SyncPanel {
	sp := &SyncPanel{configRepo: configRepo}

	sp.statusLabel = widget.NewLabel("Status: Idle")
	sp.peerLabel = widget.NewLabel("Known peers: 0")
	sp.lastSync = widget.NewLabel("Last sync: Never")

	sp.refresh()

	return sp
}

func (sp *SyncPanel) Container() fyne.CanvasObject {
	statusGroup := widget.NewCard("Sync Status", "", container.NewVBox(
		sp.statusLabel,
		sp.peerLabel,
		sp.lastSync,
	))

	infoLabel := widget.NewLabel("P2P sync is a planned feature.\nPeer discovery and sync sessions\nwill be implemented in a future update.\n\nLocal operations work fully offline.")
	infoGroup := widget.NewCard("Coming Soon", "", infoLabel)

	return container.NewVBox(statusGroup, infoGroup)
}

func (sp *SyncPanel) refresh() {
	peerID, _ := sp.configRepo.Get("peer_id")
	if peerID != "" {
		sp.peerLabel.SetText("Peer ID: " + peerID[:16] + "...")
	}
	lastSync, _ := sp.configRepo.Get("last_sync_timestamp")
	if lastSync != "" {
		sp.lastSync.SetText("Last sync: " + lastSync)
	}
}
