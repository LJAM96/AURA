package models

import "time"

type UserSubscription struct {
	ID             int           `json:"id"`
	Username       string        `json:"username"`
	CreatorID      string        `json:"creator_id"`
	ImageTypes     SelectedTypes `json:"image_types"`
	Priority       int           `json:"priority"` // 1 = highest priority, higher number = lower priority
	MediaScope     string        `json:"media_scope"`
	LibrarySection *string       `json:"library_section"`
	Enabled        bool          `json:"enabled"`
	DateCreated    time.Time     `json:"date_created"`
	DateUpdated    time.Time     `json:"date_updated"`
}
