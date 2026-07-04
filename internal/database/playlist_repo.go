package database

import (
	"database/sql"
	"time"

	"github.com/adam/y2psync/internal/model"
)

type PlaylistRepo struct {
	db *DB
}

func NewPlaylistRepo(db *DB) *PlaylistRepo {
	return &PlaylistRepo{db: db}
}

func (r *PlaylistRepo) CreateList(list *model.PlaylistList) error {
	_, err := r.db.Exec(
		`INSERT INTO playlist_lists (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		list.ID, list.Name, list.CreatedAt.UTC().Format(time.RFC3339Nano), list.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *PlaylistRepo) GetList(id string) (*model.PlaylistList, error) {
	row := r.db.QueryRow(`SELECT id, name, created_at, updated_at FROM playlist_lists WHERE id = ?`, id)
	list := &model.PlaylistList{}
	var createdStr, updatedStr string
	if err := row.Scan(&list.ID, &list.Name, &createdStr, &updatedStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	list.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	list.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return list, nil
}

func (r *PlaylistRepo) ListLists() ([]*model.PlaylistList, error) {
	rows, err := r.db.Query(`SELECT id, name, created_at, updated_at FROM playlist_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*model.PlaylistList
	for rows.Next() {
		list := &model.PlaylistList{}
		var createdStr, updatedStr string
		if err := rows.Scan(&list.ID, &list.Name, &createdStr, &updatedStr); err != nil {
			return nil, err
		}
		list.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
		list.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		lists = append(lists, list)
	}
	return lists, rows.Err()
}

func (r *PlaylistRepo) UpdateList(list *model.PlaylistList) error {
	_, err := r.db.Exec(
		`UPDATE playlist_lists SET name = ?, updated_at = ? WHERE id = ?`,
		list.Name, list.UpdatedAt.UTC().Format(time.RFC3339Nano), list.ID,
	)
	return err
}

func (r *PlaylistRepo) DeleteList(id string) error {
	_, err := r.db.Exec(`DELETE FROM playlist_lists WHERE id = ?`, id)
	return err
}

func (r *PlaylistRepo) AddEntry(entry *model.PlaylistEntry) error {
	var deletedAt *string
	if entry.DeletedAt != nil {
		s := entry.DeletedAt.UTC().Format(time.RFC3339Nano)
		deletedAt = &s
	}
	_, err := r.db.Exec(
		`INSERT INTO playlist_entries (id, playlist_list_id, youtube_video_id, display_title, created_at, sort_order, is_deleted, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.PlaylistListID, entry.YouTubeVideoID, entry.DisplayTitle, entry.CreatedAt.UTC().Format(time.RFC3339Nano), entry.SortOrder, boolToInt(entry.IsDeleted), deletedAt,
	)
	return err
}

func (r *PlaylistRepo) GetEntries(listID string) ([]*model.PlaylistEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, playlist_list_id, youtube_video_id, display_title, created_at, sort_order, is_deleted, deleted_at FROM playlist_entries WHERE playlist_list_id = ? AND is_deleted = 0 ORDER BY sort_order`,
		listID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.PlaylistEntry
	for rows.Next() {
		entry := &model.PlaylistEntry{}
		var createdStr, deletedStr *string
		var isDeleted int
		if err := rows.Scan(&entry.ID, &entry.PlaylistListID, &entry.YouTubeVideoID, &entry.DisplayTitle, &createdStr, &entry.SortOrder, &isDeleted, &deletedStr); err != nil {
			return nil, err
		}
		entry.IsDeleted = intToBool(isDeleted)
		if createdStr != nil {
			entry.CreatedAt, _ = time.Parse(time.RFC3339Nano, *createdStr)
		}
		if deletedStr != nil {
			t, _ := time.Parse(time.RFC3339Nano, *deletedStr)
			entry.DeletedAt = &t
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *PlaylistRepo) GetAllEntries() ([]*model.PlaylistEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, playlist_list_id, youtube_video_id, display_title, created_at, sort_order, is_deleted, deleted_at FROM playlist_entries ORDER BY playlist_list_id, sort_order`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.PlaylistEntry
	for rows.Next() {
		entry := &model.PlaylistEntry{}
		var createdStr, deletedStr *string
		var isDeleted int
		if err := rows.Scan(&entry.ID, &entry.PlaylistListID, &entry.YouTubeVideoID, &entry.DisplayTitle, &createdStr, &entry.SortOrder, &isDeleted, &deletedStr); err != nil {
			return nil, err
		}
		entry.IsDeleted = intToBool(isDeleted)
		if createdStr != nil {
			entry.CreatedAt, _ = time.Parse(time.RFC3339Nano, *createdStr)
		}
		if deletedStr != nil {
			t, _ := time.Parse(time.RFC3339Nano, *deletedStr)
			entry.DeletedAt = &t
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *PlaylistRepo) EntryExists(listID, youtubeVideoID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM playlist_entries WHERE playlist_list_id = ? AND youtube_video_id = ? AND is_deleted = 0`,
		listID, youtubeVideoID,
	).Scan(&count)
	return count > 0, err
}

func (r *PlaylistRepo) RemoveEntry(id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.Exec(
		`UPDATE playlist_entries SET is_deleted = 1, deleted_at = ? WHERE id = ?`,
		now, id,
	)
	return err
}

func (r *PlaylistRepo) GetMaxSortOrder(listID string) (int, error) {
	var max int
	err := r.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM playlist_entries WHERE playlist_list_id = ?`, listID).Scan(&max)
	return max, err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}
