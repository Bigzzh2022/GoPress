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

type PostSummary struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Excerpt     string       `json:"excerpt"`
	CoverImage  string       `json:"cover_image"`
	Status      string       `json:"status"`
	CategoryID  *uint        `json:"category_id"`
	Category    *CategoryDTO `json:"category"`
	Tags        []TagDTO     `json:"tags"`
	Author      UserDTO      `json:"author"`
	PublishedAt *time.Time   `json:"published_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
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
	return PostSummary{
		ID: post.ID, Title: post.Title, Slug: post.Slug, Excerpt: post.Excerpt,
		CoverImage: post.CoverImage, Status: post.Status, CategoryID: post.CategoryID,
		Category: category, Tags: tags, Author: ToUserDTO(post.Author),
		PublishedAt: post.PublishedAt, CreatedAt: post.CreatedAt, UpdatedAt: post.UpdatedAt,
	}
}

func ToPostDetail(post Post) PostDetail {
	return PostDetail{
		PostSummary: ToPostSummary(post), ContentMD: post.ContentMD, ContentHTML: post.ContentHTML,
	}
}
