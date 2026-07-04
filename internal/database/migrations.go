package database

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS playlist_lists (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS playlist_entries (
		id TEXT PRIMARY KEY,
		playlist_list_id TEXT NOT NULL,
		youtube_video_id TEXT NOT NULL,
		display_title TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0,
		is_deleted INTEGER NOT NULL DEFAULT 0,
		deleted_at TEXT,
		FOREIGN KEY (playlist_list_id) REFERENCES playlist_lists(id) ON DELETE CASCADE
	)`,
	`CREATE INDEX IF NOT EXISTS idx_playlist_entries_list_id ON playlist_entries(playlist_list_id)`,
	`CREATE INDEX IF NOT EXISTS idx_playlist_entries_video_id ON playlist_entries(youtube_video_id)`,

	`CREATE TABLE IF NOT EXISTS subscription_lists (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS subscription_entries (
		id TEXT PRIMARY KEY,
		subscription_list_id TEXT NOT NULL,
		youtube_channel_id TEXT NOT NULL,
		channel_name TEXT NOT NULL DEFAULT '',
		channel_url TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		is_deleted INTEGER NOT NULL DEFAULT 0,
		deleted_at TEXT,
		FOREIGN KEY (subscription_list_id) REFERENCES subscription_lists(id) ON DELETE CASCADE
	)`,
	`CREATE INDEX IF NOT EXISTS idx_subscription_entries_list_id ON subscription_entries(subscription_list_id)`,
	`CREATE INDEX IF NOT EXISTS idx_subscription_entries_channel_id ON subscription_entries(youtube_channel_id)`,

	`CREATE TABLE IF NOT EXISTS sync_metadata (
		peer_id TEXT PRIMARY KEY,
		last_sync_timestamp TEXT NOT NULL,
		last_sync_status TEXT NOT NULL DEFAULT ''
	)`,

	`CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`,
}
