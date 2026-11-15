package posts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"gopress/internal/model"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/database"
	"gopress/internal/platform/markdown"
	"gopress/internal/platform/textutil"
)

var (
	ErrNotFound        = errors.New("post not found")
	ErrConflict        = errors.New("post slug already exists")
	ErrInvalid         = errors.New("invalid post input")
	ErrInvalidCategory = errors.New("category does not exist")
	ErrInvalidTags     = errors.New("one or more tags do not exist")
)

type Service struct {
	db       *gorm.DB
	markdown *markdown.Renderer
	cache    *cache.Cache
}

type ListParams struct {
	Page     int
	PageSize int
	Keyword  string
	Category string
	Tag      string
	Status   string
}

type CreateInput struct {
	Title       string     `json:"title" binding:"required,max=255"`
	Slug        string     `json:"slug" binding:"omitempty,max=191"`
	Excerpt     string     `json:"excerpt"`
	ContentMD   string     `json:"content_md" binding:"required"`
	CoverImage  string     `json:"cover_image" binding:"omitempty,max=500"`
	CategoryID  *uint      `json:"category_id"`
	TagIDs      []uint     `json:"tag_ids"`
	Status      string     `json:"status" binding:"omitempty,oneof=draft published"`
	PublishedAt *time.Time `json:"published_at"`
}

type NullableUint struct {
	Set   bool
	Value *uint
}

func (n *NullableUint) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var value uint
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == 0 {
		return errors.New("category_id must be greater than zero")
	}
	n.Value = &value
	return nil
}

type UpdateInput struct {
	Title       *string      `json:"title" binding:"omitempty,max=255"`
	Slug        *string      `json:"slug" binding:"omitempty,max=191"`
	Excerpt     *string      `json:"excerpt"`
	ContentMD   *string      `json:"content_md"`
	CoverImage  *string      `json:"cover_image" binding:"omitempty,max=500"`
	CategoryID  NullableUint `json:"category_id"`
	TagIDs      []uint       `json:"tag_ids"`
	Status      *string      `json:"status" binding:"omitempty,oneof=draft published"`
	PublishedAt *time.Time   `json:"published_at"`
}

func NewService(db *gorm.DB, renderer *markdown.Renderer, publicCache *cache.Cache) *Service {
	return &Service{db: db, markdown: renderer, cache: publicCache}
}

func (s *Service) List(ctx context.Context, params ListParams, publicOnly bool) ([]model.Post, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.Post{})
	if publicOnly {
		query = published(query, time.Now())
	} else if params.Status != "" {
		query = query.Where("posts.status = ?", params.Status)
	}
	query = applyFilters(query, params)

	var total int64
	if err := query.Distinct("posts.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.Post, 0, params.PageSize)
	err := query.Select("posts.*").Scopes(withPreloads).
		Order("posts.created_at DESC").Order("posts.id DESC")
	if publicOnly {
		err = query.Select("posts.*").Scopes(withPreloads).
			Order("posts.published_at DESC").Order("posts.id DESC")
	}
	offset := (params.Page - 1) * params.PageSize
	if err := err.Offset(offset).Limit(params.PageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func applyFilters(query *gorm.DB, params ListParams) *gorm.DB {
	if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
		keyword = "%" + escapeLike(keyword) + "%"
		query = query.Where(
			"(posts.title LIKE ? OR posts.excerpt LIKE ? OR posts.content_md LIKE ?)",
			keyword, keyword, keyword,
		)
	}
	if category := strings.TrimSpace(params.Category); category != "" {
		query = query.Joins("JOIN categories ON categories.id = posts.category_id").
			Where("categories.slug = ?", category)
	}
	if tag := strings.TrimSpace(params.Tag); tag != "" {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.slug = ?", tag)
	}
	return query
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

func published(query *gorm.DB, now time.Time) *gorm.DB {
	return query.Where(
		"posts.status = ? AND posts.published_at IS NOT NULL AND posts.published_at <= ?",
		model.StatusPublished, now,
	)
}

func withPreloads(query *gorm.DB) *gorm.DB {
	return query.Preload("Author").Preload("Category").Preload("Tags", func(db *gorm.DB) *gorm.DB {
		return db.Order("tags.name ASC")
	})
}

func (s *Service) PublicDetail(ctx context.Context, slug string) (model.Post, error) {
	var post model.Post
	err := published(s.db.WithContext(ctx).Model(&model.Post{}), time.Now()).
		Where("posts.slug = ?", slug).Scopes(withPreloads).First(&post).Error
	return post, normalizeNotFound(err)
}

func (s *Service) AdminDetail(ctx context.Context, id uint) (model.Post, error) {
	return s.load(ctx, id)
}

func (s *Service) Related(ctx context.Context, slug string) ([]model.Post, error) {
	post, err := s.PublicDetail(ctx, slug)
	if err != nil {
		return nil, err
	}
	tagIDs := make([]uint, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	items := make([]model.Post, 0)
	if post.CategoryID == nil && len(tagIDs) == 0 {
		return items, nil
	}

	conditions := make([]string, 0, 2)
	arguments := make([]any, 0, 2)
	if post.CategoryID != nil {
		conditions = append(conditions, "posts.category_id = ?")
		arguments = append(arguments, *post.CategoryID)
	}
	if len(tagIDs) > 0 {
		subquery := s.db.Table("post_tags").Select("post_id").Where("tag_id IN ?", tagIDs)
		conditions = append(conditions, "posts.id IN (?)")
		arguments = append(arguments, subquery)
	}
	query := published(s.db.WithContext(ctx).Model(&model.Post{}), time.Now()).
		Where("posts.id <> ?", post.ID).Where(strings.Join(conditions, " OR "), arguments...)
	if err := query.Scopes(withPreloads).Order("posts.published_at DESC").Limit(4).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) Create(ctx context.Context, authorID uint, input CreateInput) (model.Post, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || input.ContentMD == "" {
		return model.Post{}, ErrInvalid
	}
	slug, err := s.makeSlug(ctx, title, input.Slug, 0)
	if err != nil {
		return model.Post{}, err
	}
	if err := s.validateReferences(ctx, input.CategoryID, input.TagIDs); err != nil {
		return model.Post{}, err
	}
	contentHTML, err := s.markdown.Render(input.ContentMD)
	if err != nil {
		return model.Post{}, err
	}
	status := input.Status
	if status == "" {
		status = model.StatusDraft
	}
	publishedAt := input.PublishedAt
	if status == model.StatusPublished && publishedAt == nil {
		now := time.Now()
		publishedAt = &now
	}
	excerpt := strings.TrimSpace(input.Excerpt)
	if excerpt == "" {
		excerpt = s.markdown.Excerpt(input.ContentMD, 200)
	}
	post := model.Post{
		Title: title, Slug: slug, Excerpt: excerpt, ContentMD: input.ContentMD,
		ContentHTML: contentHTML, CoverImage: strings.TrimSpace(input.CoverImage),
		Status: status, PublishedAt: publishedAt, AuthorID: authorID, CategoryID: input.CategoryID,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		if input.TagIDs != nil {
			var tags []model.Tag
			if len(input.TagIDs) > 0 {
				if err := tx.Where("id IN ?", input.TagIDs).Find(&tags).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&post).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
	if database.IsDuplicate(err) {
		return model.Post{}, ErrConflict
	}
	if err != nil {
		return model.Post{}, err
	}
	s.invalidate(ctx)
	return s.load(ctx, post.ID)
}

func (s *Service) Update(ctx context.Context, id uint, input UpdateInput) (model.Post, error) {
	post, err := s.load(ctx, id)
	if err != nil {
		return model.Post{}, err
	}
	if input.Title != nil {
		post.Title = strings.TrimSpace(*input.Title)
		if post.Title == "" {
			return model.Post{}, ErrInvalid
		}
	}
	if input.Slug != nil {
		post.Slug, err = s.makeSlug(ctx, post.Title, *input.Slug, post.ID)
		if err != nil {
			return model.Post{}, err
		}
	}
	if input.CategoryID.Set {
		if input.CategoryID.Value != nil {
			if err := s.validateReferences(ctx, input.CategoryID.Value, nil); err != nil {
				return model.Post{}, err
			}
		}
		post.CategoryID = input.CategoryID.Value
	}
	if input.TagIDs != nil {
		if err := s.validateReferences(ctx, nil, input.TagIDs); err != nil {
			return model.Post{}, err
		}
	}
	if input.ContentMD != nil {
		if *input.ContentMD == "" {
			return model.Post{}, ErrInvalid
		}
		post.ContentMD = *input.ContentMD
		post.ContentHTML, err = s.markdown.Render(post.ContentMD)
		if err != nil {
			return model.Post{}, err
		}
	}
	if input.Excerpt != nil {
		post.Excerpt = strings.TrimSpace(*input.Excerpt)
		if post.Excerpt == "" {
			post.Excerpt = s.markdown.Excerpt(post.ContentMD, 200)
		}
	}
	if input.CoverImage != nil {
		post.CoverImage = strings.TrimSpace(*input.CoverImage)
	}
	if input.Status != nil {
		post.Status = *input.Status
	}
	if input.PublishedAt != nil {
		post.PublishedAt = input.PublishedAt
	}
	if post.Status == model.StatusPublished && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Where("id = ?", post.ID).Updates(map[string]any{
			"title": post.Title, "slug": post.Slug, "excerpt": post.Excerpt,
			"content_md": post.ContentMD, "content_html": post.ContentHTML,
			"cover_image": post.CoverImage, "status": post.Status,
			"published_at": post.PublishedAt, "category_id": post.CategoryID,
		}).Error; err != nil {
			return err
		}
		if input.TagIDs != nil {
			var tags []model.Tag
			if len(input.TagIDs) > 0 {
				if err := tx.Where("id IN ?", input.TagIDs).Find(&tags).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&post).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
	if database.IsDuplicate(err) {
		return model.Post{}, ErrConflict
	}
	if err != nil {
		return model.Post{}, err
	}
	s.invalidate(ctx)
	return s.load(ctx, post.ID)
}

func (s *Service) SetStatus(ctx context.Context, id uint, status string) (model.Post, error) {
	if status != model.StatusDraft && status != model.StatusPublished {
		return model.Post{}, ErrInvalid
	}
	post, err := s.load(ctx, id)
	if err != nil {
		return model.Post{}, err
	}
	post.Status = status
	if status == model.StatusPublished && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	if err := s.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		Updates(map[string]any{"status": post.Status, "published_at": post.PublishedAt}).Error; err != nil {
		return model.Post{}, err
	}
	s.invalidate(ctx)
	return s.load(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	var post model.Post
	if err := s.db.WithContext(ctx).First(&post, id).Error; err != nil {
		return normalizeNotFound(err)
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", id).Delete(&model.PostTag{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Post{}).Error
	})
	if err == nil {
		s.invalidate(ctx)
	}
	return err
}

func (s *Service) makeSlug(ctx context.Context, title, value string, excludeID uint) (string, error) {
	if strings.TrimSpace(value) == "" {
		value = title
	}
	slug := textutil.Slugify(value)
	if slug == "" {
		return "", fmt.Errorf("%w: slug cannot be empty", ErrInvalid)
	}
	query := s.db.WithContext(ctx).Model(&model.Post{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return "", err
	}
	if count > 0 {
		return "", ErrConflict
	}
	return slug, nil
}

func (s *Service) validateReferences(ctx context.Context, categoryID *uint, tagIDs []uint) error {
	if categoryID != nil {
		if *categoryID == 0 {
			return ErrInvalidCategory
		}
		var count int64
		if err := s.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", *categoryID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrInvalidCategory
		}
	}
	if len(tagIDs) > 0 {
		var count int64
		if err := s.db.WithContext(ctx).Model(&model.Tag{}).Where("id IN ?", tagIDs).Distinct("id").Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(uniqueIDs(tagIDs))) {
			return ErrInvalidTags
		}
	}
	return nil
}

func uniqueIDs(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func (s *Service) load(ctx context.Context, id uint) (model.Post, error) {
	var post model.Post
	err := s.db.WithContext(ctx).Scopes(withPreloads).First(&post, id).Error
	return post, normalizeNotFound(err)
}

func normalizeNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *Service) invalidate(ctx context.Context) {
	_ = s.cache.BumpVersion(ctx)
}
