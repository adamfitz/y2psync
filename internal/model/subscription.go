package model

import "time"

type SubscriptionList struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionEntry struct {
	ID                 string    `json:"id"`
	SubscriptionListID string    `json:"subscription_list_id"`
	YouTubeChannelID   string    `json:"youtube_channel_id"`
	ChannelName        string    `json:"channel_name"`
	ChannelURL         string    `json:"channel_url"`
	CreatedAt          time.Time `json:"created_at"`
	IsDeleted          bool      `json:"is_deleted"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}
