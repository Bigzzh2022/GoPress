package database

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"gopress/migrations"

	"gorm.io/gorm"
)

func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(191) NOT NULL PRIMARY KEY,
		applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		var count int64
		if err := db.WithContext(ctx).Model(&struct{}{}).Table("schema_migrations").
			Where("version = ?", entry.Name()).Count(&count).Error; err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if count > 0 {
			continue
		}
		contents, err := migrations.Files.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		for _, statement := range strings.Split(string(contents), ";") {
			if strings.TrimSpace(statement) == "" {
				continue
			}
			if err := db.WithContext(ctx).Exec(statement).Error; err != nil {
				return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
			}
		}
		if err := db.WithContext(ctx).Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)", entry.Name(),
		).Error; err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
