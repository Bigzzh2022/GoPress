package model

import "time"

type UserDTO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
	Role   string `json:"role"`
}

// PublicUserDTO is what anonymous readers may see. It intentionally omits the
// email so article payloads cannot be harvested for phishing.
type PublicUserDTO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
}

type CategoryDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	PostCount   int64     `json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type TagDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PostCount int64     `json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

type SiteSettingsDTO struct {
	SiteName        string `json:"site_name"`
	SiteTagline     string `json:"site_tagline"`
	SiteDescription string `json:"site_description"`
	FooterText      string `json:"footer_text"`
	ContactEmail    string `json:"contact_email"`
}

type PostSummary struct {
	ID          uint          `json:"id"`
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	Excerpt     string        `json:"excerpt"`
	CoverImage  string        `json:"cover_image"`
	Status      string        `json:"status"`
	CategoryID  *uint         `json:"category_id"`
	Category    *CategoryDTO  `json:"category"`
	Tags        []TagDTO      `json:"tags"`
	Author      PublicUserDTO `json:"author"`
	PublishedAt *time.Time    `json:"published_at"`
	DeletedAt   *time.Time    `json:"deleted_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type PostDetail struct {
	PostSummary
	ContentMD   string `json:"content_md"`
	ContentHTML string `json:"content_html"`
}

func ToUserDTO(user User) UserDTO {
	return UserDTO{
		ID: user.ID, Name: user.Name, Email: user.Email,
		Avatar: user.Avatar, Bio: user.Bio, Role: user.Role,
	}
}

func ToPublicUserDTO(user User) PublicUserDTO {
	return PublicUserDTO{
		ID: user.ID, Name: user.Name, Avatar: user.Avatar, Bio: user.Bio,
	}
}

func ToCategoryDTO(category Category) CategoryDTO {
	return CategoryDTO{
		ID: category.ID, Name: category.Name, Slug: category.Slug,
		Description: category.Description, CreatedAt: category.CreatedAt,
	}
}

func ToTagDTO(tag Tag) TagDTO {
	return TagDTO{
		ID: tag.ID, Name: tag.Name, Slug: tag.Slug, CreatedAt: tag.CreatedAt,
	}
}

func ToSiteSettingsDTO(setting SiteSetting) SiteSettingsDTO {
	return SiteSettingsDTO{
		SiteName: setting.SiteName, SiteTagline: setting.SiteTagline,
		SiteDescription: setting.SiteDescription, FooterText: setting.FooterText,
		ContactEmail: setting.ContactEmail,
	}
}

func ToPostSummary(post Post) PostSummary {
	tags := make([]TagDTO, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tags = append(tags, ToTagDTO(tag))
	}
	var category *CategoryDTO
	if post.Category != nil {
		dto := ToCategoryDTO(*post.Category)
		category = &dto
	}
	var deletedAt *time.Time
	if post.DeletedAt.Valid {
		deletedAt = &post.DeletedAt.Time
	}
	return PostSummary{
		ID: post.ID, Title: post.Title, Slug: post.Slug, Excerpt: post.Excerpt,
		CoverImage: post.CoverImage, Status: post.Status, CategoryID: post.CategoryID,
		Category: category, Tags: tags, Author: ToPublicUserDTO(post.Author),
		PublishedAt: post.PublishedAt, DeletedAt: deletedAt,
		CreatedAt: post.CreatedAt, UpdatedAt: post.UpdatedAt,
	}
}

func ToPostDetail(post Post) PostDetail {
	return PostDetail{
		PostSummary: ToPostSummary(post), ContentMD: post.ContentMD, ContentHTML: post.ContentHTML,
	}
}
