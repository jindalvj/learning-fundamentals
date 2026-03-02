package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/tinyurl/models"
)

// Manager handles all Redis cache operations
type Manager struct {
	client     *redis.Client
	ctx        context.Context
	defaultTTL time.Duration
}

// NewManager creates a new cache manager
func NewManager(host string, port int, password string, db int) (*Manager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Manager{
		client:     client,
		ctx:        ctx,
		defaultTTL: 1 * time.Hour,
	}, nil
}

// Close closes the Redis connection
func (m *Manager) Close() error {
	return m.client.Close()
}

// Get retrieves a value from cache
func (m *Manager) Get(key string) (string, error) {
	val, err := m.client.Get(m.ctx, key).Result()
	if err == redis.Nil {
		m.client.Incr(m.ctx, "cache:misses")
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("cache get error: %w", err)
	}

	m.client.Incr(m.ctx, "cache:hits")
	return val, nil
}

// Set stores a value in cache with TTL
func (m *Manager) Set(key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = m.defaultTTL
	}

	err := m.client.Set(m.ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (m *Manager) Delete(key string) error {
	err := m.client.Del(m.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (m *Manager) Exists(key string) (bool, error) {
	count, err := m.client.Exists(m.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}

	return count > 0, nil
}

// GetURL retrieves a URL from cache by short code
func (m *Manager) GetURL(shortCode string) (string, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	return m.Get(key)
}

// SetURL stores a URL mapping in cache
func (m *Manager) SetURL(shortCode string, originalURL string, ttl time.Duration) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return m.Set(key, originalURL, ttl)
}

// DeleteURL removes a URL from cache
func (m *Manager) DeleteURL(shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	return m.Delete(key)
}

// GetStats retrieves cached statistics for a URL
func (m *Manager) GetStats(shortCode string) (*models.URLStats, error) {
	key := fmt.Sprintf("stats:%s", shortCode)
	val, err := m.Get(key)
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, nil
	}

	var stats models.URLStats
	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	return &stats, nil
}

// SetStats stores statistics in cache
func (m *Manager) SetStats(shortCode string, stats *models.URLStats, ttl time.Duration) error {
	key := fmt.Sprintf("stats:%s", shortCode)

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	if ttl == 0 {
		ttl = 5 * time.Minute // Default 5 minutes for stats
	}

	return m.Set(key, string(data), ttl)
}

// Increment atomically increments a counter
func (m *Manager) Increment(key string) (int64, error) {
	val, err := m.client.Incr(m.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("cache increment error: %w", err)
	}

	return val, nil
}

// Decrement atomically decrements a counter
func (m *Manager) Decrement(key string) (int64, error) {
	val, err := m.client.Decr(m.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("cache decrement error: %w", err)
	}

	return val, nil
}

// GetMany retrieves multiple keys at once
func (m *Manager) GetMany(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	values, err := m.client.MGet(m.ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("cache mget error: %w", err)
	}

	result := make(map[string]string)
	for i, val := range values {
		if val != nil {
			if strVal, ok := val.(string); ok {
				result[keys[i]] = strVal
			}
		}
	}

	return result, nil
}

// SetMany stores multiple key-value pairs at once
func (m *Manager) SetMany(mapping map[string]string, ttl time.Duration) error {
	if len(mapping) == 0 {
		return nil
	}

	pipe := m.client.Pipeline()

	for key, value := range mapping {
		if ttl > 0 {
			pipe.Set(m.ctx, key, value, ttl)
		} else {
			pipe.Set(m.ctx, key, value, 0)
		}
	}

	_, err := pipe.Exec(m.ctx)
	if err != nil {
		return fmt.Errorf("cache mset error: %w", err)
	}

	return nil
}

// DeleteMany deletes multiple keys at once
func (m *Manager) DeleteMany(keys []string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	count, err := m.client.Del(m.ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("cache delete many error: %w", err)
	}

	return count, nil
}

// FlushAll clears all keys from the current database
func (m *Manager) FlushAll() error {
	err := m.client.FlushDB(m.ctx).Err()
	if err != nil {
		return fmt.Errorf("cache flush error: %w", err)
	}

	return nil
}

// GetCacheStats retrieves cache performance statistics
func (m *Manager) GetCacheStats() (*models.CacheStats, error) {
	// Get hits and misses
	hitsStr, err := m.client.Get(m.ctx, "cache:hits").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	hits, _ := strconv.ParseInt(hitsStr, 10, 64)

	missesStr, err := m.client.Get(m.ctx, "cache:misses").Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	misses, _ := strconv.ParseInt(missesStr, 10, 64)

	// Calculate hit rate
	total := hits + misses
	hitRate := "0.00%"
	if total > 0 {
		hitRate = fmt.Sprintf("%.2f%%", float64(hits)/float64(total)*100)
	}

	// Get Redis info
	_, err = m.client.Info(m.ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	// Get memory info
	memInfo, err := m.client.Info(m.ctx, "memory").Result()
	if err != nil {
		return nil, err
	}

	// Get total keys
	totalKeys, err := m.client.DBSize(m.ctx).Result()
	if err != nil {
		return nil, err
	}

	// Parse connected clients (simplified)
	connectedClients := 0

	stats := &models.CacheStats{
		Hits:       hits,
		Misses:     misses,
		HitRate:    hitRate,
		TotalKeys:  totalKeys,
		UsedMemory: parseMemoryInfo(memInfo),
		TotalConns: connectedClients,
	}

	return stats, nil
}

// ResetStats resets cache statistics counters
func (m *Manager) ResetStats() error {
	err := m.client.Set(m.ctx, "cache:hits", 0, 0).Err()
	if err != nil {
		return err
	}

	err = m.client.Set(m.ctx, "cache:misses", 0, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

// Ping checks if Redis is responsive
func (m *Manager) Ping() error {
	return m.client.Ping(m.ctx).Err()
}

// Helper function to parse memory info from Redis INFO command
func parseMemoryInfo(info string) string {
	// Simplified parsing - in production, parse the actual INFO output
	return "N/A"
}
