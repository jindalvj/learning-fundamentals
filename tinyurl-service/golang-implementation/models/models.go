package models

import (
	"time"
)

// URL represents a shortened URL entry
type URL struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	UserID      *int64     `json:"user_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// URLStats represents access statistics for a URL
type URLStats struct {
	ID           int64      `json:"id"`
	ShortCode    string     `json:"short_code"`
	AccessCount  int64      `json:"access_count"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
}

// Click represents a detailed click event
type Click struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	ClickedAt   time.Time `json:"clicked_at"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	UserAgent   *string   `json:"user_agent,omitempty"`
	Referrer    *string   `json:"referrer,omitempty"`
	CountryCode *string   `json:"country_code,omitempty"`
}

// CreateURLRequest represents the request to create a short URL
type CreateURLRequest struct {
	URL           string `json:"url" binding:"required"`
	CustomAlias   string `json:"custom_alias,omitempty"`
	ExpiresInDays *int   `json:"expires_in_days,omitempty"`
	UserID        *int64 `json:"user_id,omitempty"`
}

// CreateURLResponse represents the response after creating a short URL
type CreateURLResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	IsExisting  bool      `json:"is_existing,omitempty"`
}

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// URLInfo represents complete information about a URL
type URLInfo struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	Stats       *URLStats  `json:"stats,omitempty"`
}

// ServiceStats represents overall service statistics
type ServiceStats struct {
	Database DatabaseStats `json:"database"`
	Cache    CacheStats    `json:"cache"`
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	ActiveURLs       int64 `json:"active_urls"`
	InactiveURLs     int64 `json:"inactive_urls"`
	TotalClicks      int64 `json:"total_clicks"`
	TotalClickEvents int64 `json:"total_click_events"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64  `json:"hits"`
	Misses      int64  `json:"misses"`
	HitRate     string `json:"hit_rate"`
	TotalKeys   int64  `json:"total_keys"`
	UsedMemory  string `json:"used_memory"`
	TotalConns  int    `json:"connected_clients"`
}
