package database

import (
	"fmt"
	"time"

	"github.com/adam/y2psync/internal/model"
	"github.com/google/uuid"
)

func (db *DB) RestoreFrom(backupDB *DB, fresh bool) error {
	srcPlaylistRepo := NewPlaylistRepo(backupDB)
	srcSubRepo := NewSubscriptionRepo(backupDB)
	dstPlaylistRepo := NewPlaylistRepo(db)
	dstSubRepo := NewSubscriptionRepo(db)

	srcPlaylists, err := srcPlaylistRepo.ListLists()
	if err != nil {
		return fmt.Errorf("read source playlists: %w", err)
	}

	srcPlaylistEntries := make(map[string][]*model.PlaylistEntry, len(srcPlaylists))
	for _, pl := range srcPlaylists {
		entries, err := srcPlaylistRepo.GetEntries(pl.ID)
		if err != nil {
			return fmt.Errorf("read source entries for %s: %w", pl.Name, err)
		}
		srcPlaylistEntries[pl.ID] = entries
	}

	srcSubLists, err := srcSubRepo.ListLists()
	if err != nil {
		return fmt.Errorf("read source subscription lists: %w", err)
	}

	srcSubEntries := make(map[string][]*model.SubscriptionEntry, len(srcSubLists))
	for _, sl := range srcSubLists {
		entries, err := srcSubRepo.GetEntries(sl.ID)
		if err != nil {
			return fmt.Errorf("read source sub entries for %s: %w", sl.Name, err)
		}
		srcSubEntries[sl.ID] = entries
	}

	if fresh {
		return db.restoreFresh(dstPlaylistRepo, dstSubRepo, srcPlaylists, srcPlaylistEntries, srcSubLists, srcSubEntries, backupDB)
	}
	return db.restoreMerge(dstPlaylistRepo, dstSubRepo, srcPlaylists, srcPlaylistEntries, srcSubLists, srcSubEntries)
}

func (db *DB) restoreFresh(
	dstPlaylist *PlaylistRepo, dstSub *SubscriptionRepo,
	srcPlaylists []*model.PlaylistList, srcPlaylistEntries map[string][]*model.PlaylistEntry,
	srcSubLists []*model.SubscriptionList, srcSubEntries map[string][]*model.SubscriptionEntry,
	backupDB *DB,
) error {
	if _, err := db.Exec("DELETE FROM playlist_entries"); err != nil {
		return fmt.Errorf("clear playlist entries: %w", err)
	}
	if _, err := db.Exec("DELETE FROM playlist_lists"); err != nil {
		return fmt.Errorf("clear playlist lists: %w", err)
	}
	if _, err := db.Exec("DELETE FROM subscription_entries"); err != nil {
		return fmt.Errorf("clear subscription entries: %w", err)
	}
	if _, err := db.Exec("DELETE FROM subscription_lists"); err != nil {
		return fmt.Errorf("clear subscription lists: %w", err)
	}

	for _, pl := range srcPlaylists {
		if err := dstPlaylist.CreateList(pl); err != nil {
			return fmt.Errorf("restore playlist %s: %w", pl.Name, err)
		}
		for _, e := range srcPlaylistEntries[pl.ID] {
			if err := dstPlaylist.AddEntry(e); err != nil {
				return fmt.Errorf("restore playlist entry %s: %w", e.YouTubeVideoID, err)
			}
		}
	}

	for _, sl := range srcSubLists {
		if err := dstSub.CreateList(sl); err != nil {
			return fmt.Errorf("restore sub list %s: %w", sl.Name, err)
		}
		for _, e := range srcSubEntries[sl.ID] {
			if err := dstSub.AddEntry(e); err != nil {
				return fmt.Errorf("restore sub entry %s: %w", e.YouTubeChannelID, err)
			}
		}
	}

	backupConfig := NewConfigRepo(backupDB)
	dstConfig := NewConfigRepo(db)

	for _, key := range []string{"peer_id", "master_sync_key_salt", "sync_group_key", "rendezvous_tag", "sync_key_configured", "last_sync_timestamp"} {
		val, err := backupConfig.Get(key)
		if err != nil || val == "" {
			continue
		}
		dstConfig.Set(key, val)
	}

	return nil
}

func (db *DB) restoreMerge(
	dstPlaylist *PlaylistRepo, dstSub *SubscriptionRepo,
	srcPlaylists []*model.PlaylistList, srcPlaylistEntries map[string][]*model.PlaylistEntry,
	srcSubLists []*model.SubscriptionList, srcSubEntries map[string][]*model.SubscriptionEntry,
) error {
	existingPlaylists, err := dstPlaylist.ListLists()
	if err != nil {
		return fmt.Errorf("read existing playlists: %w", err)
	}
	playlistByName := make(map[string]*model.PlaylistList, len(existingPlaylists))
	for _, pl := range existingPlaylists {
		playlistByName[pl.Name] = pl
	}

	for _, pl := range srcPlaylists {
		existing, found := playlistByName[pl.Name]
		if !found {
			pl.ID = uuid.New().String()
			pl.CreatedAt = time.Now().UTC()
			pl.UpdatedAt = time.Now().UTC()
			if err := dstPlaylist.CreateList(pl); err != nil {
				return fmt.Errorf("create playlist %s: %w", pl.Name, err)
			}
			existing = pl
		}

		maxOrder, _ := dstPlaylist.GetMaxSortOrder(existing.ID)
		for _, e := range srcPlaylistEntries[pl.ID] {
			exists, _ := dstPlaylist.EntryExists(existing.ID, e.YouTubeVideoID)
			if exists {
				continue
			}
			maxOrder++
			e.ID = uuid.New().String()
			e.PlaylistListID = existing.ID
			e.SortOrder = maxOrder
			e.CreatedAt = time.Now().UTC()
			if err := dstPlaylist.AddEntry(e); err != nil {
				return fmt.Errorf("add entry %s: %w", e.YouTubeVideoID, err)
			}
		}
	}

	existingSubs, err := dstSub.ListLists()
	if err != nil {
		return fmt.Errorf("read existing subscription lists: %w", err)
	}
	subByName := make(map[string]*model.SubscriptionList, len(existingSubs))
	for _, sl := range existingSubs {
		subByName[sl.Name] = sl
	}

	for _, sl := range srcSubLists {
		existing, found := subByName[sl.Name]
		if !found {
			sl.ID = uuid.New().String()
			sl.CreatedAt = time.Now().UTC()
			sl.UpdatedAt = time.Now().UTC()
			if err := dstSub.CreateList(sl); err != nil {
				return fmt.Errorf("create sub list %s: %w", sl.Name, err)
			}
			existing = sl
		}

		for _, e := range srcSubEntries[sl.ID] {
			exists, _ := dstSub.EntryExists(existing.ID, e.YouTubeChannelID)
			if exists {
				continue
			}
			e.ID = uuid.New().String()
			e.SubscriptionListID = existing.ID
			e.CreatedAt = time.Now().UTC()
			if err := dstSub.AddEntry(e); err != nil {
				return fmt.Errorf("add sub entry %s: %w", e.YouTubeChannelID, err)
			}
		}
	}

	return nil
}
