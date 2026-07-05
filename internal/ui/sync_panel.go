package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/adam/y2psync/internal/sync"
)

type SyncPanel struct {
	syncer *sync.Syncer

	statusLabel *widget.Label
	knownLabel  *widget.Label
	syncedLabel *widget.Label
	lastSync    *widget.Label
}

func NewSyncPanel(syncer *sync.Syncer) *SyncPanel {
	sp := &SyncPanel{
		syncer: syncer,
	}

	ss := syncer.SyncStatus()
	sp.statusLabel = widget.NewLabel(fmt.Sprintf("Status: %s", ss.State))
	sp.knownLabel = widget.NewLabel(fmt.Sprintf("Devices known: %d", ss.KnownPeers))
	sp.syncedLabel = widget.NewLabel(fmt.Sprintf("Devices synced: %d", ss.SyncedPeers))
	lastSync := ss.LastSync
	if lastSync == "" {
		lastSync = "Never"
	}
	sp.lastSync = widget.NewLabel("Last sync: " + lastSync)

	go sp.watchStatus()

	return sp
}

func (sp *SyncPanel) Container() fyne.CanvasObject {
	statusGroup := widget.NewCard("Sync Status", "", container.NewVBox(
		sp.statusLabel,
		sp.knownLabel,
		sp.syncedLabel,
		sp.lastSync,
	))

	infoGroup := widget.NewCard("P2P Sync", "",
		widget.NewLabel("Sync runs automatically when a Master Sync Key\nis configured in Settings.\n\nDiscovers other y2psync devices via the\npublic IPFS DHT. Connections are encrypted\nvia libp2p Noise.\n\nStatus becomes 'Settled' once all known\ndevices have been found."))

	return container.NewVBox(statusGroup, infoGroup)
}

func (sp *SyncPanel) watchStatus() {
	ch := sp.syncer.SyncStatusChan()
	for ss := range ch {
		fyne.Do(func() {
			sp.statusLabel.SetText(fmt.Sprintf("Status: %s", ss.State))
			sp.knownLabel.SetText(fmt.Sprintf("Devices known: %d", ss.KnownPeers))
			sp.syncedLabel.SetText(fmt.Sprintf("Devices synced: %d", ss.SyncedPeers))
			lastSync := ss.LastSync
			if lastSync == "" {
				lastSync = "Never"
			}
			sp.lastSync.SetText("Last sync: " + lastSync)
		})
	}
}
