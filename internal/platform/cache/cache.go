package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const versionKey = "gopress:public:version"

type Cache struct {
	Client *redis.Client
	TTL    time.Duration
}

func New(addr, password string, db int) *Cache {
	return &Cache{
		Client: redis.NewClient(&redis.Options{
			Addr: addr, Password: password, DB: db,
			DialTimeout: 500 * time.Millisecond, ReadTimeout: 500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond, MaxRetries: 1,
		}),
		TTL: 5 * time.Minute,
	}
}

func (c *Cache) PublicKey(ctx context.Context, resource, parameters string) string {
	version := uint64(0)
	if c != nil && c.Client != nil {
		if value, err := c.Client.Get(ctx, versionKey).Uint64(); err == nil {
			version = value
		}
	}
	digest := sha256.Sum256([]byte(parameters))
	return fmt.Sprintf("gopress:public:v%d:%s:%s", version, resource, hex.EncodeToString(digest[:8]))
}

func (c *Cache) GetJSON(ctx context.Context, key string, destination any) bool {
	if c == nil || c.Client == nil {
		return false
	}
	value, err := c.Client.Get(ctx, key).Bytes()
	if err != nil || json.Unmarshal(value, destination) != nil {
		return false
	}
	return true
}

func (c *Cache) SetJSON(ctx context.Context, key string, value any) {
	if c == nil || c.Client == nil {
		return
	}
	encoded, err := json.Marshal(value)
	if err == nil {
		_ = c.Client.Set(ctx, key, encoded, c.TTL).Err()
	}
}

func (c *Cache) BumpVersion(ctx context.Context) error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Incr(ctx, versionKey).Err()
}
