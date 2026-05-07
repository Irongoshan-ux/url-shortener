// Package model defines the URL entity exchanged between repository, service, and JSON APIs.
package model

import "time"

// URL is a shortened link record. UserID and IsDeleted are omitted from JSON responses where encoded.
type URL struct {
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"-"`
	IsDeleted   bool      `json:"-"`
}
