package database

import (
	"time"
)

type APIKeyAttr struct {
	AccountName string
	PublicKey   []byte
	SecretKey   []byte
}

type UsersAttrs struct {
	Username string
	Name     string
	IsAdmin  bool
}

type CategoriesAttrs struct {
	Id          uint64
	Slug        string
	Name        string
	Description string
	Sort        int
}

type TagAttrs struct {
	Id   uint64
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
	UUID   string
	PostID uint64
	UserID uint64
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
