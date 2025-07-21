package payload

import (
	"github.com/google/uuid"
	"time"
)

type Posts struct {
	Uuid          uuid.NullUUID `json:"uuid"`
	Slug          string        `json:"slug"`
	Title         string        `json:"title"`
	Excerpt       string        `json:"excerpt"`
	Content       string        `json:"content"`
	CoverImageURL string        `json:"cover_image_url"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
