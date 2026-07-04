package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"github.com/adam/y2psync/internal/database"
	"github.com/adam/y2psync/internal/sync"
	"github.com/adam/y2psync/internal/ui"
)

func main() {
	dataDirFlag := flag.String("data-dir", "", "path to data directory (default: ~/.y2psync)")
	flag.Parse()

	dataDir := *dataDirFlag
	if dataDir == "" {
		dataDir = defaultDataDir()
	}

	db, err := database.Open(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	playlistRepo := database.NewPlaylistRepo(db)
	subRepo := database.NewSubscriptionRepo(db)
	configRepo := database.NewConfigRepo(db)

	syncer := sync.NewSyncer(db, configRepo)

	a := app.NewWithID("com.y2psync.app")
	win := a.NewWindow("y2psync")

	ui.NewApp(win, db, playlistRepo, subRepo, configRepo, syncer)

	if syncer.IsSyncConfigured() {
		go syncer.Run()
	}

	win.ShowAndRun()
	syncer.Stop()
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "y2psync-data")
	}
	return filepath.Join(home, ".y2psync")
}
