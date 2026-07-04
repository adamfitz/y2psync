package database

import (
	"database/sql"
	"time"

	"github.com/adam/y2psync/internal/model"
)

type SubscriptionRepo struct {
	db *DB
}

func NewSubscriptionRepo(db *DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) CreateList(list *model.SubscriptionList) error {
	_, err := r.db.Exec(
		`INSERT INTO subscription_lists (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		list.ID, list.Name, list.CreatedAt.UTC().Format(time.RFC3339Nano), list.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *SubscriptionRepo) GetList(id string) (*model.SubscriptionList, error) {
	row := r.db.QueryRow(`SELECT id, name, created_at, updated_at FROM subscription_lists WHERE id = ?`, id)
	list := &model.SubscriptionList{}
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

func (r *SubscriptionRepo) ListLists() ([]*model.SubscriptionList, error) {
	rows, err := r.db.Query(`SELECT id, name, created_at, updated_at FROM subscription_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*model.SubscriptionList
	for rows.Next() {
		list := &model.SubscriptionList{}
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

func (r *SubscriptionRepo) UpdateList(list *model.SubscriptionList) error {
	_, err := r.db.Exec(
		`UPDATE subscription_lists SET name = ?, updated_at = ? WHERE id = ?`,
		list.Name, list.UpdatedAt.UTC().Format(time.RFC3339Nano), list.ID,
	)
	return err
}

func (r *SubscriptionRepo) DeleteList(id string) error {
	_, err := r.db.Exec(`DELETE FROM subscription_lists WHERE id = ?`, id)
	return err
}

func (r *SubscriptionRepo) AddEntry(entry *model.SubscriptionEntry) error {
	var deletedAt *string
	if entry.DeletedAt != nil {
		s := entry.DeletedAt.UTC().Format(time.RFC3339Nano)
		deletedAt = &s
	}
	_, err := r.db.Exec(
		`INSERT INTO subscription_entries (id, subscription_list_id, youtube_channel_id, channel_name, channel_url, created_at, is_deleted, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.SubscriptionListID, entry.YouTubeChannelID, entry.ChannelName, entry.ChannelURL, entry.CreatedAt.UTC().Format(time.RFC3339Nano), boolToInt(entry.IsDeleted), deletedAt,
	)
	return err
}

func (r *SubscriptionRepo) GetEntries(listID string) ([]*model.SubscriptionEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, subscription_list_id, youtube_channel_id, channel_name, channel_url, created_at, is_deleted, deleted_at FROM subscription_entries WHERE subscription_list_id = ? AND is_deleted = 0`,
		listID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.SubscriptionEntry
	for rows.Next() {
		entry := &model.SubscriptionEntry{}
		var createdStr, deletedStr *string
		var isDeleted int
		if err := rows.Scan(&entry.ID, &entry.SubscriptionListID, &entry.YouTubeChannelID, &entry.ChannelName, &entry.ChannelURL, &createdStr, &isDeleted, &deletedStr); err != nil {
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

func (r *SubscriptionRepo) GetAllEntries() ([]*model.SubscriptionEntry, error) {
	rows, err := r.db.Query(
		`SELECT id, subscription_list_id, youtube_channel_id, channel_name, channel_url, created_at, is_deleted, deleted_at FROM subscription_entries`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*model.SubscriptionEntry
	for rows.Next() {
		entry := &model.SubscriptionEntry{}
		var createdStr, deletedStr *string
		var isDeleted int
		if err := rows.Scan(&entry.ID, &entry.SubscriptionListID, &entry.YouTubeChannelID, &entry.ChannelName, &entry.ChannelURL, &createdStr, &isDeleted, &deletedStr); err != nil {
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

func (r *SubscriptionRepo) EntryExists(listID, youtubeChannelID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM subscription_entries WHERE subscription_list_id = ? AND youtube_channel_id = ? AND is_deleted = 0`,
		listID, youtubeChannelID,
	).Scan(&count)
	return count > 0, err
}

func (r *SubscriptionRepo) RemoveEntry(id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.db.Exec(
		`UPDATE subscription_entries SET is_deleted = 1, deleted_at = ? WHERE id = ?`,
		now, id,
	)
	return err
}
