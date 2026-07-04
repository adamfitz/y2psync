package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/adam/y2psync/internal/database"
	"github.com/adam/y2psync/internal/sync"
)

type SyncPanel struct {
	configRepo *database.ConfigRepo
	syncer     *sync.Syncer
	syncMu     chan struct{}

	statusLabel *widget.Label
	peerLabel   *widget.Label
	lastSync    *widget.Label
	toggleBtn   *widget.Button
}

func NewSyncPanel(configRepo *database.ConfigRepo, syncer *sync.Syncer) *SyncPanel {
	sp := &SyncPanel{
		configRepo: configRepo,
		syncer:     syncer,
		syncMu:     make(chan struct{}, 1),
	}

	sp.statusLabel = widget.NewLabel(fmt.Sprintf("Status: %s", syncer.Status()))
	sp.peerLabel = widget.NewLabel("Known peers: 0")
	sp.lastSync = widget.NewLabel("Last sync: Never")

	sp.toggleBtn = widget.NewButton("Start Sync", func() {
		sp.toggleSync()
	})

	sp.refresh()

	go sp.watchStatus()

	return sp
}

func (sp *SyncPanel) Container() fyne.CanvasObject {
	statusGroup := widget.NewCard("Sync Status", "", container.NewVBox(
		sp.statusLabel,
		sp.peerLabel,
		sp.lastSync,
		sp.toggleBtn,
	))

	infoGroup := widget.NewCard("P2P Sync", "",
		widget.NewLabel("Discovers other y2psync devices via the\npublic IPFS DHT using your Master Sync Key.\n\nConnections are encrypted via libp2p Noise.\nRequires an internet connection."))

	return container.NewVBox(statusGroup, infoGroup)
}

func (sp *SyncPanel) watchStatus() {
	ch := sp.syncer.StatusChan()
	for range ch {
		sp.refresh()
	}
}

func (sp *SyncPanel) refresh() {
	st := sp.syncer.Status()
	sp.statusLabel.SetText(fmt.Sprintf("Status: %s", st))
	sp.peerLabel.SetText(fmt.Sprintf("Peers synced: %d", sp.syncer.PeerCount()))

	lastSync, _ := sp.configRepo.Get("last_sync_timestamp")
	if lastSync != "" {
		sp.lastSync.SetText("Last sync: " + lastSync)
	}

	if st == sync.StatusRunning {
		sp.toggleBtn.SetText("Stop Sync")
	} else {
		sp.toggleBtn.SetText("Start Sync")
	}
}

func (sp *SyncPanel) toggleSync() {
	select {
	case sp.syncMu <- struct{}{}:
	default:
		return
	}
	defer func() { <-sp.syncMu }()

	if sp.syncer.Status() == sync.StatusRunning {
		sp.syncer.Stop()
		return
	}

	if !sp.syncer.IsSyncConfigured() {
		dialog.ShowInformation("Sync Not Configured",
			"Go to Settings and configure a Master Sync Key first.",
			fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	go func() {
		if err := sp.syncer.Start(); err != nil {
			dialog.ShowError(fmt.Errorf("start sync: %w", err), fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}()
	sp.refresh()
}
