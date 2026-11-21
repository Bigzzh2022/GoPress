package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopress/internal/config"
	httpapi "gopress/internal/http"
	"gopress/internal/platform/cache"
	"gopress/internal/platform/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("create upload directory %s: %v", cfg.UploadDir, err)
	}
	startup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(startup, cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	if err := database.Migrate(startup, db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	publicCache := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer publicCache.Client.Close()
	if err := publicCache.Client.Ping(startup).Err(); err != nil {
		log.Printf("redis unavailable; continuing with database and fail-open rate limiting: %v", err)
	}
	router := httpapi.NewRouter(cfg, db, publicCache)
	server := &http.Server{
		Addr: ":" + cfg.Port, Handler: router,
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("GoPress API listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve HTTP: %v", err)
		}
	case <-signals:
		shutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}
}
