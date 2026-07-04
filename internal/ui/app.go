package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/adam/y2psync/internal/database"
	"github.com/adam/y2psync/internal/sync"
)

type App struct {
	win          fyne.Window
	playlistView *PlaylistView
	subView      *SubscriptionView
	syncPanel    *SyncPanel
	settingsView *SettingsView

	playlistRepo *database.PlaylistRepo
	subRepo      *database.SubscriptionRepo
	configRepo   *database.ConfigRepo
	db           *database.DB
	syncer       *sync.Syncer
}

func NewApp(
	win fyne.Window,
	db *database.DB,
	playlistRepo *database.PlaylistRepo,
	subRepo *database.SubscriptionRepo,
	configRepo *database.ConfigRepo,
	syncer *sync.Syncer,
) *App {
	a := &App{
		win:          win,
		db:           db,
		playlistRepo: playlistRepo,
		subRepo:      subRepo,
		configRepo:   configRepo,
		syncer:       syncer,
	}

	a.playlistView = NewPlaylistView(playlistRepo, win)
	a.subView = NewSubscriptionView(subRepo, win)
	a.syncPanel = NewSyncPanel(syncer)
	a.settingsView = NewSettingsView(db, configRepo, win, syncer,
		a.playlistView.refreshList,
		a.subView.refreshEntries,
	)

	playlistTab := container.NewTabItemWithIcon("Playlists", theme.ListIcon(), a.playlistView.Container())
	subTab := container.NewTabItemWithIcon("Subscriptions", theme.MailComposeIcon(), a.subView.Container())
	syncTab := container.NewTabItemWithIcon("Sync", theme.ComputerIcon(), a.syncPanel.Container())
	settingsTab := container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), a.settingsView.Container())

	tabs := container.NewAppTabs(playlistTab, subTab, syncTab, settingsTab)
	tabs.SetTabLocation(container.TabLocationTop)

	win.SetContent(tabs)
	win.Resize(fyne.NewSize(900, 600))

	return a
}
