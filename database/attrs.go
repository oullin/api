package database

import (
	"time"
)

type UsersAttrs struct {
	Username string
	Name     string
	IsAdmin  bool
}

type CategoriesAttrs struct {
	Slug        string
	Name        string
	Description string
}

type TagAttrs struct {
	Slug string
	Name string
}

type CommentsAttrs struct {
	UUID       string
	PostID     uint64
	AuthorID   uint64
	ParentID   *uint64
	Content    string
	ApprovedAt *time.Time
}

type LikesAttrs struct {
	UUID   string `gorm:"type:uuid;unique;not null"`
	PostID uint64 `gorm:"not null;index;uniqueIndex:idx_likes_post_user"`
	UserID uint64 `gorm:"not null;index;uniqueIndex:idx_likes_post_user"`
}

type NewsletterAttrs struct {
	FirstName      string
	LastName       string
	Email          string
	SubscribedAt   *time.Time
	UnsubscribedAt *time.Time
}

type PostViewsAttr struct {
	Post      Post
	User      User
	IPAddress string
	UserAgent string
}

type PostsAttrs struct {
	AuthorID    uint64
	Slug        string
	Title       string
	Excerpt     string
	Content     string
	ImageURL    string
	PublishedAt *time.Time
	Categories  []CategoriesAttrs
	Tags        []TagAttrs
}
