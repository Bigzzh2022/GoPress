package model

import "time"

const (
	RoleAdmin       = "admin"
	StatusDraft     = "draft"
	StatusPublished = "published"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	Name         string    `gorm:"size:100;not null" json:"-"`
	Email        string    `gorm:"size:191;not null;uniqueIndex" json:"-"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Avatar       string    `gorm:"size:500" json:"-"`
	Bio          string    `gorm:"type:text" json:"-"`
	Role         string    `gorm:"size:20;not null;default:admin;index" json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	Name        string    `gorm:"size:100;not null" json:"-"`
	Slug        string    `gorm:"size:191;not null;uniqueIndex" json:"-"`
	Description string    `gorm:"type:text" json:"-"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	Name      string    `gorm:"size:100;not null" json:"-"`
	Slug      string    `gorm:"size:191;not null;uniqueIndex" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Post struct {
	ID          uint       `gorm:"primaryKey" json:"-"`
	Title       string     `gorm:"size:255;not null" json:"-"`
	Slug        string     `gorm:"size:191;not null;uniqueIndex" json:"-"`
	Excerpt     string     `gorm:"type:text" json:"-"`
	ContentMD   string     `gorm:"type:longtext;not null" json:"-"`
	ContentHTML string     `gorm:"type:longtext;not null" json:"-"`
	CoverImage  string     `gorm:"size:500" json:"-"`
	Status      string     `gorm:"size:20;not null;default:draft;index:idx_posts_status_published,priority:1" json:"-"`
	PublishedAt *time.Time `gorm:"index:idx_posts_status_published,priority:2" json:"-"`
	AuthorID    uint       `gorm:"not null;index" json:"-"`
	Author      User       `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CategoryID  *uint      `gorm:"index" json:"-"`
	Category    *Category  `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Tags        []Tag      `gorm:"many2many:post_tags;joinForeignKey:PostID;joinReferences:TagID" json:"-"`
	CreatedAt   time.Time  `gorm:"index" json:"-"`
	UpdatedAt   time.Time  `json:"-"`
}

type PostTag struct {
	PostID    uint      `gorm:"primaryKey" json:"-"`
	TagID     uint      `gorm:"primaryKey;index" json:"-"`
	CreatedAt time.Time `json:"-"`
}

func (PostTag) TableName() string {
	return "post_tags"
}
