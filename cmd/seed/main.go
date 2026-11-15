package main

import (
	"context"
	"log"
	"time"

	"gopress/internal/config"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/database"
	"gopress/internal/platform/markdown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer database.Close(db)
	if err := database.Migrate(ctx, db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	publicCache := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer publicCache.Client.Close()
	if err := database.Seed(ctx, db, cfg, markdown.NewRenderer(), publicCache); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	log.Printf("seed completed for %s (demo=%t)", cfg.AdminEmail, cfg.SeedDemo)
}
