package model

import "time"

type PlaylistList struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlaylistEntry struct {
	ID              string    `json:"id"`
	PlaylistListID  string    `json:"playlist_list_id"`
	YouTubeVideoID  string    `json:"youtube_video_id"`
	DisplayTitle    string    `json:"display_title"`
	CreatedAt       time.Time `json:"created_at"`
	SortOrder       int       `json:"sort_order"`
	IsDeleted       bool      `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
