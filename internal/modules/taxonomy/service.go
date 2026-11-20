package taxonomy

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"gopress/internal/model"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/database"
	"gopress/internal/platform/textutil"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrTagNotFound      = errors.New("tag not found")
	ErrConflict         = errors.New("slug already exists")
	ErrInvalid          = errors.New("invalid taxonomy input")
)

type Service struct {
	db    *gorm.DB
	cache *cache.Cache
}

type CategoryInput struct {
	Name        string `json:"name" binding:"required,max=100"`
	Slug        string `json:"slug" binding:"omitempty,max=191"`
	Description string `json:"description"`
}

type TagInput struct {
	Name string `json:"name" binding:"required,max=100"`
	Slug string `json:"slug" binding:"omitempty,max=191"`
}

type countRow struct {
	ID    uint  `gorm:"column:id"`
	Count int64 `gorm:"column:count"`
}

func NewService(db *gorm.DB, publicCache *cache.Cache) *Service {
	return &Service{db: db, cache: publicCache}
}

func (s *Service) ListCategories(ctx context.Context, publicOnly bool) ([]model.CategoryDTO, error) {
	var categories []model.Category
	if err := s.db.WithContext(ctx).Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64)
	query := s.db.WithContext(ctx).Table("posts").
		Select("category_id AS id, COUNT(*) AS count").
		Where("category_id IS NOT NULL AND deleted_at IS NULL")
	if publicOnly {
		query = publicPosts(query)
	}
	var rows []countRow
	if err := query.Group("category_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ID] = row.Count
	}
	items := make([]model.CategoryDTO, 0, len(categories))
	for _, category := range categories {
		dto := model.ToCategoryDTO(category)
		dto.PostCount = counts[category.ID]
		items = append(items, dto)
	}
	return items, nil
}

func (s *Service) ListTags(ctx context.Context, publicOnly bool) ([]model.TagDTO, error) {
	var tags []model.Tag
	if err := s.db.WithContext(ctx).Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint]int64)
	query := s.db.WithContext(ctx).Table("post_tags").
		Select("post_tags.tag_id AS id, COUNT(DISTINCT posts.id) AS count").
		Joins("JOIN posts ON posts.id = post_tags.post_id").
		Where("posts.deleted_at IS NULL")
	if publicOnly {
		query = publicPosts(query)
	}
	var rows []countRow
	if err := query.Group("post_tags.tag_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ID] = row.Count
	}
	items := make([]model.TagDTO, 0, len(tags))
	for _, tag := range tags {
		dto := model.ToTagDTO(tag)
		dto.PostCount = counts[tag.ID]
		items = append(items, dto)
	}
	return items, nil
}

func publicPosts(query *gorm.DB) *gorm.DB {
	return query.Where(
		"posts.status = ? AND posts.published_at IS NOT NULL AND posts.published_at <= ? AND posts.deleted_at IS NULL",
		model.StatusPublished, time.Now(),
	)
}

func (s *Service) CreateCategory(ctx context.Context, input CategoryInput) (model.CategoryDTO, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.CategoryDTO{}, ErrInvalid
	}
	slug, err := s.uniqueSlug(ctx, "categories", name, input.Slug, 0)
	if err != nil {
		return model.CategoryDTO{}, err
	}
	category := model.Category{Name: name, Slug: slug, Description: strings.TrimSpace(input.Description)}
	if err := s.db.WithContext(ctx).Create(&category).Error; err != nil {
		if database.IsDuplicate(err) {
			return model.CategoryDTO{}, ErrConflict
		}
		return model.CategoryDTO{}, err
	}
	s.invalidate(ctx)
	return model.ToCategoryDTO(category), nil
}

func (s *Service) UpdateCategory(ctx context.Context, id uint, input CategoryInput) (model.CategoryDTO, error) {
	var category model.Category
	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.CategoryDTO{}, ErrCategoryNotFound
		}
		return model.CategoryDTO{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.CategoryDTO{}, ErrInvalid
	}
	slug, err := s.uniqueSlug(ctx, "categories", name, input.Slug, id)
	if err != nil {
		return model.CategoryDTO{}, err
	}
	category.Name = name
	category.Slug = slug
	category.Description = strings.TrimSpace(input.Description)
	if err := s.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).
		Updates(map[string]any{"name": category.Name, "slug": category.Slug, "description": category.Description}).Error; err != nil {
		if database.IsDuplicate(err) {
			return model.CategoryDTO{}, ErrConflict
		}
		return model.CategoryDTO{}, err
	}
	s.invalidate(ctx)
	return model.ToCategoryDTO(category), nil
}

func (s *Service) DeleteCategory(ctx context.Context, id uint) error {
	var category model.Category
	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Where("category_id = ?", id).Update("category_id", nil).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Category{}).Error
	})
	if err == nil {
		s.invalidate(ctx)
	}
	return err
}

func (s *Service) CreateTag(ctx context.Context, input TagInput) (model.TagDTO, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.TagDTO{}, ErrInvalid
	}
	slug, err := s.uniqueSlug(ctx, "tags", name, input.Slug, 0)
	if err != nil {
		return model.TagDTO{}, err
	}
	tag := model.Tag{Name: name, Slug: slug}
	if err := s.db.WithContext(ctx).Create(&tag).Error; err != nil {
		if database.IsDuplicate(err) {
			return model.TagDTO{}, ErrConflict
		}
		return model.TagDTO{}, err
	}
	s.invalidate(ctx)
	return model.ToTagDTO(tag), nil
}

func (s *Service) UpdateTag(ctx context.Context, id uint, input TagInput) (model.TagDTO, error) {
	var tag model.Tag
	if err := s.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.TagDTO{}, ErrTagNotFound
		}
		return model.TagDTO{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.TagDTO{}, ErrInvalid
	}
	slug, err := s.uniqueSlug(ctx, "tags", name, input.Slug, id)
	if err != nil {
		return model.TagDTO{}, err
	}
	tag.Name = name
	tag.Slug = slug
	if err := s.db.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", id).
		Updates(map[string]any{"name": tag.Name, "slug": tag.Slug}).Error; err != nil {
		if database.IsDuplicate(err) {
			return model.TagDTO{}, ErrConflict
		}
		return model.TagDTO{}, err
	}
	s.invalidate(ctx)
	return model.ToTagDTO(tag), nil
}

func (s *Service) DeleteTag(ctx context.Context, id uint) error {
	var tag model.Tag
	if err := s.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", id).Delete(&model.PostTag{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Tag{}).Error
	})
	if err == nil {
		s.invalidate(ctx)
	}
	return err
}

func (s *Service) uniqueSlug(ctx context.Context, table, name, value string, excludeID uint) (string, error) {
	if strings.TrimSpace(value) == "" {
		value = name
	}
	slug := textutil.Slugify(value)
	if slug == "" {
		return "", ErrInvalid
	}
	query := s.db.WithContext(ctx).Table(table).Where("slug = ?", slug)
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

func (s *Service) invalidate(ctx context.Context) {
	_ = s.cache.BumpVersion(ctx)
}
