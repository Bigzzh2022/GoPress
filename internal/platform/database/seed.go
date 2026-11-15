package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gopress/internal/config"
	"gopress/internal/model"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/markdown"
)

const demoPublishedMarkdown = `# Building a Production-Ready Go Blog

GoPress combines a small API with a pleasant reading experience. This article demonstrates the Markdown features supported by the backend.

## Architecture

The request flows through **Gin**, domain services, and **GORM** before reaching MySQL. Redis provides short-lived, versioned cache entries.

> Keep rendering at the trust boundary: Markdown is converted and sanitized on the server.

### API example

` + "```go\n" + `package main

import "fmt"

func main() {
	fmt.Println("Hello, GoPress!")
}
` + "```\n\n" + `The publishing checklist is intentionally short:

- Write in Markdown
- Add a category and descriptive tags
- Preview sanitized HTML
- Publish when the article is ready

| Component | Responsibility |
| --- | --- |
| MySQL | Durable content |
| Redis | Sessions and cache |
| Goldmark | Markdown parsing |

Read the [Go documentation](https://go.dev/doc/) for more information.
`

const demoDraftMarkdown = `# Editorial Workflow Notes

This draft is visible only through the admin API.

## Ideas

1. Explain cache invalidation.
2. Include realistic API examples.
3. Review the reading experience on mobile.

` + "```sql\n" + `SELECT title, published_at
FROM posts
WHERE status = 'draft';
` + "```\n"

func Seed(ctx context.Context, db *gorm.DB, cfg config.Config, renderer *markdown.Renderer, publicCache *cache.Cache) error {
	var admin model.User
	err := db.WithContext(ctx).Where("role = ?", model.RoleAdmin).Order("id ASC").First(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if len(cfg.AdminPassword) < 8 {
			return fmt.Errorf("ADMIN_PASSWORD must contain at least 8 characters")
		}
		email := strings.ToLower(strings.TrimSpace(cfg.AdminEmail))
		if email == "" || strings.TrimSpace(cfg.AdminName) == "" {
			return fmt.Errorf("ADMIN_NAME and ADMIN_EMAIL cannot be empty")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}
		admin = model.User{
			Name: strings.TrimSpace(cfg.AdminName), Email: email,
			PasswordHash: string(hash), Role: model.RoleAdmin,
		}
		if err := db.WithContext(ctx).Create(&admin).Error; err != nil {
			return fmt.Errorf("create admin: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("find admin: %w", err)
	}

	if cfg.SeedDemo {
		if err := seedDemo(ctx, db, admin, renderer); err != nil {
			return err
		}
	}
	return publicCache.BumpVersion(ctx)
}

func seedDemo(ctx context.Context, db *gorm.DB, admin model.User, renderer *markdown.Renderer) error {
	backend, err := findOrCreateCategory(ctx, db, "Backend Engineering", "backend", "Go, databases, and API engineering.")
	if err != nil {
		return err
	}
	publishing, err := findOrCreateCategory(ctx, db, "Publishing", "publishing", "Editorial workflows and content design.")
	if err != nil {
		return err
	}
	goTag, err := findOrCreateTag(ctx, db, "Go", "go")
	if err != nil {
		return err
	}
	gormTag, err := findOrCreateTag(ctx, db, "GORM", "gorm")
	if err != nil {
		return err
	}
	mysqlTag, err := findOrCreateTag(ctx, db, "MySQL", "mysql")
	if err != nil {
		return err
	}
	markdownTag, err := findOrCreateTag(ctx, db, "Markdown", "markdown")
	if err != nil {
		return err
	}

	if err := createDemoPost(ctx, db, admin, backend, []model.Tag{goTag, gormTag, mysqlTag, markdownTag}, renderer,
		"Building a Production-Ready Go Blog", "production-ready-go-blog", demoPublishedMarkdown,
		model.StatusPublished, time.Now().Add(-24*time.Hour)); err != nil {
		return err
	}
	return createDemoPost(ctx, db, admin, publishing, []model.Tag{markdownTag}, renderer,
		"Editorial Workflow Notes", "editorial-workflow-notes", demoDraftMarkdown,
		model.StatusDraft, time.Time{})
}

func findOrCreateCategory(ctx context.Context, db *gorm.DB, name, slug, description string) (model.Category, error) {
	var category model.Category
	err := db.WithContext(ctx).Where("slug = ?", slug).Attrs(model.Category{
		Name: name, Slug: slug, Description: description,
	}).FirstOrCreate(&category).Error
	return category, err
}

func findOrCreateTag(ctx context.Context, db *gorm.DB, name, slug string) (model.Tag, error) {
	var tag model.Tag
	err := db.WithContext(ctx).Where("slug = ?", slug).Attrs(model.Tag{Name: name, Slug: slug}).FirstOrCreate(&tag).Error
	return tag, err
}

func createDemoPost(ctx context.Context, db *gorm.DB, author model.User, category model.Category, tags []model.Tag,
	renderer *markdown.Renderer, title, slug, source, status string, publishedAt time.Time) error {
	var existing model.Post
	err := db.WithContext(ctx).Where("slug = ?", slug).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	html, err := renderer.Render(source)
	if err != nil {
		return err
	}
	post := model.Post{
		Title: title, Slug: slug, Excerpt: renderer.Excerpt(source, 200),
		ContentMD: source, ContentHTML: html, Status: status,
		AuthorID: author.ID, CategoryID: &category.ID,
	}
	if status == model.StatusPublished {
		post.PublishedAt = &publishedAt
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		return tx.Model(&post).Association("Tags").Replace(tags)
	})
}
