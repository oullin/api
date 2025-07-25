package posts

import (
	"time"
)

type IndexRequestBody struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
	Tag      string `json:"tag"`
	Text     string `json:"text"`
}

type PostResponse struct {
	UUID          string     `json:"uuid"`
	Author        UserData   `json:"author"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Excerpt       string     `json:"excerpt"`
	Content       string     `json:"content"`
	CoverImageURL string     `json:"cover_image_url"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Associations
	Categories []CategoryData `json:"categories"`
	Tags       []TagData      `json:"tags"`
}

type UserData struct {
	UUID              string    `json:"uuid"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Username          string    `json:"username"`
	DisplayName       string    `json:"display_name"`
	Bio               string    `json:"bio"`
	PictureFileName   string    `json:"picture_file_name"`
	ProfilePictureURL string    `json:"profile_picture_url"`
	IsAdmin           bool      `json:"is_admin"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CategoryData struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TagData struct {
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
