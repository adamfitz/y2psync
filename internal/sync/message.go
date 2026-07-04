package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/adam/y2psync/internal/database"
	"github.com/adam/y2psync/internal/model"
	"github.com/google/uuid"
)

const ProtocolID = "/y2psync/1.0.0"

type SyncPlaylist struct {
	Name    string      `json:"name"`
	Entries []SyncEntry `json:"entries"`
}

type SyncEntry struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`
}

type SyncSub struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelURL  string `json:"channel_url"`
}

type SyncMessage struct {
	Type          string         `json:"type"`
	PeerID        string         `json:"peer_id,omitempty"`
	Playlists     []SyncPlaylist `json:"playlists,omitempty"`
	Subscriptions []SyncSub      `json:"subscriptions,omitempty"`
}

func encodeJSONMessage(w io.Writer, msg *SyncMessage) error {
	return json.NewEncoder(w).Encode(msg)
}

func decodeJSONMessage(r io.Reader) (*SyncMessage, error) {
	var msg SyncMessage
	if err := json.NewDecoder(r).Decode(&msg); err != nil {
		return nil, fmt.Errorf("decode sync message: %w", err)
	}
	return &msg, nil
}

func collectLocalData(srcPlaylist *database.PlaylistRepo, srcSub *database.SubscriptionRepo) (*SyncMessage, error) {
	playlists, err := srcPlaylist.ListLists()
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}

	syncPlaylists := make([]SyncPlaylist, 0, len(playlists))
	for _, pl := range playlists {
		entries, err := srcPlaylist.GetEntries(pl.ID)
		if err != nil {
			return nil, fmt.Errorf("get entries: %w", err)
		}
		syncEntries := make([]SyncEntry, 0, len(entries))
		for _, e := range entries {
			syncEntries = append(syncEntries, SyncEntry{
				VideoID: e.YouTubeVideoID,
				Title:   e.DisplayTitle,
			})
		}
		syncPlaylists = append(syncPlaylists, SyncPlaylist{
			Name:    pl.Name,
			Entries: syncEntries,
		})
	}

	subs, err := srcSub.ListLists()
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	syncSubs := make([]SyncSub, 0)
	for _, sl := range subs {
		entries, err := srcSub.GetEntries(sl.ID)
		if err != nil {
			return nil, fmt.Errorf("get sub entries: %w", err)
		}
		for _, e := range entries {
			syncSubs = append(syncSubs, SyncSub{
				ChannelID:   e.YouTubeChannelID,
				ChannelName: e.ChannelName,
				ChannelURL:  e.ChannelURL,
			})
		}
	}

	return &SyncMessage{
		Type:          "data",
		Playlists:     syncPlaylists,
		Subscriptions: syncSubs,
	}, nil
}

func MergeIncomingData(dst *database.DB, msg *SyncMessage) error {
	dstPlaylist := database.NewPlaylistRepo(dst)
	dstSub := database.NewSubscriptionRepo(dst)

	for _, sp := range msg.Playlists {
		lists, err := dstPlaylist.ListLists()
		if err != nil {
			return err
		}

		var existingID string
		for _, l := range lists {
			if l.Name == sp.Name {
				existingID = l.ID
				break
			}
		}

		if existingID == "" {
			id := uuid.New().String()
			if err := dstPlaylist.CreateList(&model.PlaylistList{
				ID:        id,
				Name:      sp.Name,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}); err != nil {
				return err
			}
			existingID = id
		}

		maxOrder, _ := dstPlaylist.GetMaxSortOrder(existingID)
		for _, e := range sp.Entries {
			exists, _ := dstPlaylist.EntryExists(existingID, e.VideoID)
			if exists {
				continue
			}
			maxOrder++
			if err := dstPlaylist.AddEntry(&model.PlaylistEntry{
				ID:             uuid.New().String(),
				PlaylistListID: existingID,
				YouTubeVideoID: e.VideoID,
				DisplayTitle:   e.Title,
				CreatedAt:      time.Now().UTC(),
				SortOrder:      maxOrder,
			}); err != nil {
				return err
			}
		}
	}

	for _, ss := range msg.Subscriptions {
		lists, err := dstSub.ListLists()
		if err != nil {
			return err
		}

		var existingID string
		defaultName := "My Subscriptions"
		for _, l := range lists {
			if l.Name == defaultName {
				existingID = l.ID
				break
			}
		}
		if existingID == "" {
			id := uuid.New().String()
			if err := dstSub.CreateList(&model.SubscriptionList{
				ID:        id,
				Name:      defaultName,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}); err != nil {
				return err
			}
			existingID = id
		}

		exists, _ := dstSub.EntryExists(existingID, ss.ChannelID)
		if exists {
			continue
		}
		if err := dstSub.AddEntry(&model.SubscriptionEntry{
			ID:                 uuid.New().String(),
			SubscriptionListID: existingID,
			YouTubeChannelID:   ss.ChannelID,
			ChannelName:        ss.ChannelName,
			ChannelURL:         ss.ChannelURL,
			CreatedAt:          time.Now().UTC(),
		}); err != nil {
			return err
		}
	}

	return nil
}
