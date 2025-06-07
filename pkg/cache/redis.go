package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache wraps the Redis client
// Used as an alternative to a DB when in DB-less mode.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis client instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,     // e.g., "redis:6379"
		Password: password, // empty string if no password
		DB:       db,       // 0 is default
	})

	return &RedisCache{
		client: rdb,
		ctx:    ctx,
	}
}

// SetMessage caches a JSON-RPC message with a TTL
func (r *RedisCache) SetMessage(key string, msg any, ttl time.Duration) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, ttl).Err()
}

// GetMessage retrieves a cached message by key
func (r *RedisCache) GetMessage(key string) (*any, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var msg any
	if err := json.Unmarshal([]byte(val), &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteMessage removes a cached entry by key
func (r *RedisCache) DeleteMessage(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// example usage
//
// func main() {
// 	cache := db.NewRedisCache("redis:6379", "", 0)

// 	msg := db.RPCMessage{
// 		ID:     "abc123",
// 		Method: "getStatus",
// 		Params: map[string]interface{}{"check": "ping"},
// 		Result: map[string]interface{}{"status": "ok"},
// 	}

// 	key := "rpc:abc123"

// 	if err := cache.SetMessage(key, msg, 5*time.Minute); err != nil {
// 		log.Fatal(err)
// 	}

// 	cached, err := cache.GetMessage(key)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("🔁 Cached Message: %+v\n", cached)
// }
