package database

import (
	"gorm.io/gorm"
	"slices"
	"time"
)

const DriverName = "postgres"

var schemaTables = []string{
	"users", "posts", "categories",
	"post_categories", "tags", "post_tags",
	"post_views", "post_views", "comments",
	"likes", "newsletters", "api_keys",
}

func GetSchemaTables() []string {
	return schemaTables
}

func isValidTable(seed string) bool {
	return slices.Contains(schemaTables, seed)
}

type APIKey struct {
	ID          int64  `gorm:"primaryKey"`
	UUID        string `gorm:"type:uuid;unique;not null"`
	AccountName string `gorm:"column:account_name;size:50;not null;unique;uniqueIndex:uq_account_keys"`
	PublicKey   string `gorm:"column:public_key;size:50;not null;unique;index;uniqueIndex:uq_account_keys"`
	SecretKey   string `gorm:"column:secret_key;size:50;not null;unique;index;uniqueIndex:uq_account_keys"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type User struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement"`
	UUID              string         `gorm:"type:uuid;unique;not null"`
	FirstName         string         `gorm:"type:varchar(250);not null"`
	LastName          string         `gorm:"type:varchar(250);not null"`
	Username          string         `gorm:"type:varchar(50);unique;not null"`
	DisplayName       string         `gorm:"type:varchar(255)"`
	Email             string         `gorm:"type:varchar(250);unique;not null"`
	PasswordHash      string         `gorm:"type:varchar(255);not null"`
	PublicToken       string         `gorm:"type:varchar(250);not null;index:idx_public_token"`
	Bio               string         `gorm:"type:text"`
	PictureFileName   string         `gorm:"type:varchar(2048)"`
	ProfilePictureURL string         `gorm:"type:varchar(2048)"`
	IsAdmin           bool           `gorm:"default:false"`
	VerifiedAt        time.Time      `gorm:"index:idx_users_verified_at"`
	CreatedAt         time.Time      `gorm:"default:CURRENT_TIMESTAMP;index:idx_users_created_at"`
	UpdatedAt         time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt         gorm.DeletedAt `gorm:"index:idx_users_deleted_at"`

	// Associations
	Posts     []Post     `gorm:"foreignKey:AuthorID"`
	Comments  []Comment  `gorm:"foreignKey:AuthorID"`
	PostViews []PostView `gorm:"foreignKey:UserID"`
	Likes     []Like     `gorm:"foreignKey:UserID"`
}

type Post struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement"`
	UUID          string     `gorm:"type:uuid;unique;not null"`
	AuthorID      uint64     `gorm:"not null;index:idx_posts_author_id"`
	Author        User       `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
	Slug          string     `gorm:"type:varchar(255);unique;not null"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Excerpt       string     `gorm:"type:text;not null"`
	Content       string     `gorm:"type:text;not null"`
	CoverImageURL string     `gorm:"type:varchar(2048)"`
	PublishedAt   *time.Time `gorm:"index:idx_posts_published_at"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt     gorm.DeletedAt

	// Associations
	Categories []Category `gorm:"many2many:post_categories;"`
	Tags       []Tag      `gorm:"many2many:post_tags;"`
	PostViews  []PostView `gorm:"foreignKey:PostID"`
	Comments   []Comment  `gorm:"foreignKey:PostID"`
	Likes      []Like     `gorm:"foreignKey:PostID"`
}

type Category struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	UUID        string    `gorm:"type:uuid;unique;not null"`
	Name        string    `gorm:"type:varchar(255);unique;not null"`
	Slug        string    `gorm:"type:varchar(255);unique;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt

	// Associations
	Posts []Post `gorm:"many2many:post_categories;"`
}

type PostCategory struct {
	PostID     uint64    `gorm:"primaryKey"`
	CategoryID uint64    `gorm:"primaryKey"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Tag struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	UUID        string    `gorm:"type:uuid;unique;not null"`
	Name        string    `gorm:"type:varchar(255);unique;not null"`
	Slug        string    `gorm:"type:varchar(255);unique;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt

	// Associations
	Posts []Post `gorm:"many2many:post_tags;"`
}

type PostTag struct {
	PostID    uint64    `gorm:"primaryKey"`
	TagID     uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type PostView struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	PostID    uint64    `gorm:"not null;index:idx_post_views_post_viewed_at"`
	Post      Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	UserID    *uint64   `gorm:"index"` // Can be NULL for anonymous views
	User      *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	IPAddress string    `gorm:"type:inet"`
	UserAgent string    `gorm:"type:text"`
	ViewedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;index:idx_post_views_post_viewed_at"`
}

type Comment struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	UUID       string    `gorm:"type:uuid;unique;not null"`
	PostID     uint64    `gorm:"not null;index:idx_comments_post_id"`
	Post       Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	AuthorID   uint64    `gorm:"not null;index"`
	Author     User      `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
	ParentID   *uint64   `gorm:"index"` // For nested comments
	Parent     *Comment  `gorm:"foreignKey:ParentID"`
	Replies    []Comment `gorm:"foreignKey:ParentID"`
	Content    string    `gorm:"type:text;not null"`
	ApprovedAt *time.Time
	CreatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP;index:idx_comments_post_created_at"`
	UpdatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type Like struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UUID      string    `gorm:"type:uuid;unique;not null"`
	PostID    uint64    `gorm:"not null;index;uniqueIndex:idx_likes_post_user"`
	Post      Post      `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	UserID    uint64    `gorm:"not null;index;uniqueIndex:idx_likes_post_user"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt
}

type Newsletter struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement"`
	FirstName      string     `gorm:"type:varchar(250);not null"`
	LastName       string     `gorm:"type:varchar(250);not null"`
	Email          string     `gorm:"type:varchar(250);unique;not null"`
	SubscribedAt   *time.Time `gorm:"index:idx_newsletters_subscribed_at;type:datetime"`
	UnsubscribedAt *time.Time `gorm:"index:idx_newsletters_unsubscribed_at;type:datetime"`
	CreatedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_newsletters_created_at"`
	UpdatedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
}
