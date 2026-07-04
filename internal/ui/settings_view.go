package ui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/adam/y2psync/internal/crypto"
	"github.com/adam/y2psync/internal/database"
)

type SettingsView struct {
	win          fyne.Window
	db           *database.DB
	configRepo   *database.ConfigRepo
	peerIDLabel  *widget.Label
	dbPathLabel  *widget.Label
	syncKeyEntry *widget.Entry
	saveKeyBtn   *widget.Button
	clearKeyBtn  *widget.Button
}

func NewSettingsView(db *database.DB, configRepo *database.ConfigRepo, win fyne.Window) *SettingsView {
	sv := &SettingsView{db: db, configRepo: configRepo, win: win}

	sv.peerIDLabel = widget.NewLabel("Loading...")
	sv.syncKeyEntry = widget.NewPasswordEntry()
	sv.syncKeyEntry.SetPlaceHolder("Enter Master Sync Key (min 12 chars)")

	sv.saveKeyBtn = widget.NewButton("Save Sync Key", func() {
		sv.saveSyncKey()
	})

	sv.clearKeyBtn = widget.NewButton("Clear Sync Key", func() {
		sv.clearSyncKey()
	})

	sv.ensurePeerID()

	return sv
}

func (sv *SettingsView) Container() fyne.CanvasObject {
	backupBtn := widget.NewButton("Backup Database...", func() {
		sv.showBackupDialog()
	})

	dbPath := sv.db.Path()
	sv.dbPathLabel = widget.NewLabel(dbPath)

	databaseGroup := widget.NewCard("Database", "", container.NewVBox(
		container.NewBorder(nil, nil, widget.NewLabel("Location:"), nil, sv.dbPathLabel),
		container.NewHBox(backupBtn),
	))

	identityGroup := widget.NewCard("Device Identity", "", container.NewVBox(
		sv.peerIDLabel,
	))

	syncGroup := widget.NewCard("Master Sync Key", "", container.NewVBox(
		widget.NewLabel("Set a passphrase to enable P2P sync between devices."),
		sv.syncKeyEntry,
		container.NewHBox(sv.saveKeyBtn, sv.clearKeyBtn),
	))

	return container.NewVBox(databaseGroup, identityGroup, syncGroup)
}

func (sv *SettingsView) ensurePeerID() {
	peerID, err := sv.configRepo.Get("peer_id")
	if err != nil || peerID == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			sv.peerIDLabel.SetText("Error generating Peer ID")
			return
		}
		peerID = hex.EncodeToString(b)
		if err := sv.configRepo.Set("peer_id", peerID); err != nil {
			sv.peerIDLabel.SetText("Error saving Peer ID")
			return
		}
	}
	sv.peerIDLabel.SetText(fmt.Sprintf("Peer ID: %s", peerID))
}

func (sv *SettingsView) saveSyncKey() {
	key := sv.syncKeyEntry.Text
	if len(key) < 12 {
		dialog.ShowError(fmt.Errorf("Master Sync Key must be at least 12 characters"), sv.win)
		return
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	syncGroupKey := crypto.DeriveSyncGroupKey(key, salt)
	rendezvousTag := crypto.DeriveRendezvousTag(key)

	if err := sv.configRepo.Set("master_sync_key_salt", hex.EncodeToString(salt)); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if err := sv.configRepo.Set("sync_group_key", hex.EncodeToString(syncGroupKey)); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if err := sv.configRepo.Set("rendezvous_tag", rendezvousTag); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if err := sv.configRepo.Set("sync_key_configured", "true"); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}
	if err := sv.configRepo.Set("last_sync_timestamp", time.Now().UTC().Format(time.RFC3339)); err != nil {
		dialog.ShowError(err, sv.win)
		return
	}

	dialog.ShowInformation("Sync Key Saved", "Master Sync Key has been configured.\n\nSync Group Key and Rendezvous Tag derived.\nSync will activate when other devices are found.", sv.win)
}

func (sv *SettingsView) clearSyncKey() {
	dialog.ShowConfirm("Clear Sync Key",
		"This will disable sync. Other devices will no longer be able to sync with this device. Local data will be preserved.",
		func(ok bool) {
			if !ok {
				return
			}
			sv.configRepo.Delete("master_sync_key_salt")
			sv.configRepo.Delete("sync_group_key")
			sv.configRepo.Delete("rendezvous_tag")
			sv.configRepo.Delete("sync_key_configured")
			dialog.ShowInformation("Cleared", "Sync key cleared. Local data preserved.", sv.win)
		}, sv.win)
}

func (sv *SettingsView) showBackupDialog() {
	save := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("save dialog error: %w", err), sv.win)
			return
		}
		if writer == nil {
			return
		}
		destPath := writer.URI().Path()
		writer.Close()

		if err := sv.db.BackupTo(destPath); err != nil {
			dialog.ShowError(fmt.Errorf("backup failed: %w", err), sv.win)
			return
		}
		dialog.ShowInformation("Backup Complete",
			fmt.Sprintf("Database backed up to:\n%s", destPath), sv.win)
	}, sv.win)
	save.Resize(fyne.NewSize(640, 480))
	save.Show()
}
