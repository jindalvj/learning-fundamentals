package service

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yourusername/tinyurl/cache"
	"github.com/yourusername/tinyurl/database"
	"github.com/yourusername/tinyurl/encoder"
	"github.com/yourusername/tinyurl/models"
)

// URLShortener is the core service for URL shortening operations
type URLShortener struct {
	db                  *database.Manager
	cache               *cache.Manager
	encoder             *encoder.Base62Encoder
	baseURL             string
	cacheTTL            time.Duration
	statsCacheTTL       time.Duration
	enableDeduplication bool
	minShortCodeLength  int
	urlRegex            *regexp.Regexp
}

// Config holds configuration for the URL shortener service
type Config struct {
	BaseURL             string
	CacheTTL            time.Duration
	StatsCacheTTL       time.Duration
	EnableDeduplication bool
	MinShortCodeLength  int
}

// NewURLShortener creates a new URL shortener service
func NewURLShortener(db *database.Manager, cache *cache.Manager, config Config) *URLShortener {
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}
	if config.StatsCacheTTL == 0 {
		config.StatsCacheTTL = 5 * time.Minute
	}
	if config.MinShortCodeLength == 0 {
		config.MinShortCodeLength = 6
	}

	// URL validation regex
	urlRegex := regexp.MustCompile(`^https?://(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+[A-Z]{2,6}\.?|localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?::\d+)?(?:/?|[/?]\S+)$`)

	return &URLShortener{
		db:                  db,
		cache:               cache,
		encoder:             encoder.NewBase62Encoder(),
		baseURL:             strings.TrimSuffix(config.BaseURL, "/"),
		cacheTTL:            config.CacheTTL,
		statsCacheTTL:       config.StatsCacheTTL,
		enableDeduplication: config.EnableDeduplication,
		minShortCodeLength:  config.MinShortCodeLength,
		urlRegex:            urlRegex,
	}
}

// IsValidURL validates URL format
func (s *URLShortener) IsValidURL(urlStr string) bool {
	if len(urlStr) == 0 || len(urlStr) > 2048 {
		return false
	}

	// Parse URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	return s.urlRegex.MatchString(strings.ToUpper(urlStr))
}

// IsAliasValid validates custom alias format
func (s *URLShortener) IsAliasValid(alias string) bool {
	if len(alias) < 4 || len(alias) > 10 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, alias)
	return matched
}

// SanitizeURL sanitizes and normalizes URL
func (s *URLShortener) SanitizeURL(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	return urlStr
}

// CreateShortURL creates a shortened URL
func (s *URLShortener) CreateShortURL(req models.CreateURLRequest) (*models.CreateURLResponse, error) {
	// Sanitize and validate
	originalURL := s.SanitizeURL(req.URL)

	if !s.IsValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format: %s", originalURL)
	}

	var shortCode string
	var urlID int64
	var err error

	// Handle custom alias
	if req.CustomAlias != "" {
		if !s.IsAliasValid(req.CustomAlias) {
			return nil, fmt.Errorf("invalid alias format: %s", req.CustomAlias)
		}

		available, err := s.db.CheckAliasAvailable(req.CustomAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check alias availability: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("alias '%s' is already taken", req.CustomAlias)
		}

		shortCode = req.CustomAlias

		// Insert with custom alias
		urlID, err = s.db.InsertURL(originalURL, shortCode, req.UserID, req.ExpiresInDays)
		if err != nil {
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}
	} else {
		// Check for existing URL (deduplication)
		if s.enableDeduplication {
			existing, err := s.db.GetURLByOriginalURL(originalURL)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing URL: %w", err)
			}
			if existing != nil {
				return &models.CreateURLResponse{
					ShortCode:   existing.ShortCode,
					OriginalURL: existing.OriginalURL,
					ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existing.ShortCode),
					CreatedAt:   existing.CreatedAt,
					IsExisting:  true,
				}, nil
			}
		}

		// Insert new URL with empty short code initially
		urlID, err = s.db.InsertURL(originalURL, "", req.UserID, req.ExpiresInDays)
		if err != nil {
			return nil, fmt.Errorf("failed to insert URL: %w", err)
		}

		// Generate short code from ID
		shortCode = s.encoder.GenerateShortCode(urlID, s.minShortCodeLength)

		// Update the record with the short code
		if err := s.db.UpdateShortCode(urlID, shortCode); err != nil {
			return nil, fmt.Errorf("failed to update short code: %w", err)
		}
	}

	// Initialize statistics
	if err := s.db.InitializeStats(shortCode); err != nil {
		return nil, fmt.Errorf("failed to initialize stats: %w", err)
	}

	// Pre-warm cache
	_ = s.cache.SetURL(shortCode, originalURL, s.cacheTTL)

	return &models.CreateURLResponse{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		CreatedAt:   time.Now(),
		IsExisting:  false,
	}, nil
}

// GetOriginalURL retrieves original URL from short code
func (s *URLShortener) GetOriginalURL(shortCode string, trackClick bool, ipAddress, userAgent, referrer *string) (string, error) {
	// Check cache first
	cachedURL, err := s.cache.GetURL(shortCode)
	if err != nil {
		return "", fmt.Errorf("cache error: %w", err)
	}

	if cachedURL != "" {
		// Cache hit - async increment stats
		if trackClick {
			go s.incrementStatsAsync(shortCode, ipAddress, userAgent, referrer)
		}
		return cachedURL, nil
	}

	// Cache miss - query database
	urlData, err := s.db.GetURLByShortCode(shortCode)
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	if urlData == nil {
		return "", fmt.Errorf("short code not found")
	}

	// Check if active
	if !urlData.IsActive {
		return "", fmt.Errorf("URL is inactive")
	}

	// Check if expired
	if urlData.ExpiresAt != nil && urlData.ExpiresAt.Before(time.Now()) {
		// Mark as inactive
		_ = s.db.DeactivateURL(shortCode)
		return "", fmt.Errorf("URL has expired")
	}

	originalURL := urlData.OriginalURL

	// Update cache
	_ = s.cache.SetURL(shortCode, originalURL, s.cacheTTL)

	// Async increment stats
	if trackClick {
		go s.incrementStatsAsync(shortCode, ipAddress, userAgent, referrer)
	}

	return originalURL, nil
}

// incrementStatsAsync increments access statistics asynchronously
func (s *URLShortener) incrementStatsAsync(shortCode string, ipAddress, userAgent, referrer *string) {
	// Increment counter in database
	if err := s.db.IncrementAccessCount(shortCode); err != nil {
		fmt.Printf("Error incrementing stats for %s: %v\n", shortCode, err)
	}

	// Optionally log detailed click event
	if ipAddress != nil || userAgent != nil || referrer != nil {
		if err := s.db.LogClick(shortCode, ipAddress, userAgent, referrer); err != nil {
			fmt.Printf("Error logging click for %s: %v\n", shortCode, err)
		}
	}

	// Invalidate stats cache
	statsKey := fmt.Sprintf("stats:%s", shortCode)
	_ = s.cache.Delete(statsKey)
}

// GetStats retrieves statistics for a short URL
func (s *URLShortener) GetStats(shortCode string) (*models.URLStats, error) {
	// Check cache first
	cachedStats, err := s.cache.GetStats(shortCode)
	if err != nil {
		return nil, fmt.Errorf("cache error: %w", err)
	}

	if cachedStats != nil {
		return cachedStats, nil
	}

	// Cache miss - query database
	stats, err := s.db.GetStats(shortCode)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if stats != nil {
		// Cache for future requests
		_ = s.cache.SetStats(shortCode, stats, s.statsCacheTTL)
	}

	return stats, nil
}

// GetURLInfo retrieves complete information about a short URL
func (s *URLShortener) GetURLInfo(shortCode string) (*models.URLInfo, error) {
	// Get URL data
	urlData, err := s.db.GetURLByShortCode(shortCode)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if urlData == nil {
		return nil, fmt.Errorf("short code not found")
	}

	// Get stats
	stats, _ := s.GetStats(shortCode)

	info := &models.URLInfo{
		ShortCode:   urlData.ShortCode,
		OriginalURL: urlData.OriginalURL,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, urlData.ShortCode),
		CreatedAt:   urlData.CreatedAt,
		ExpiresAt:   urlData.ExpiresAt,
		IsActive:    urlData.IsActive,
		Stats:       stats,
	}

	return info, nil
}

// DeleteURL soft deletes a URL
func (s *URLShortener) DeleteURL(shortCode string) error {
	// Deactivate in database
	if err := s.db.DeactivateURL(shortCode); err != nil {
		return fmt.Errorf("failed to deactivate URL: %w", err)
	}

	// Remove from cache
	_ = s.cache.DeleteURL(shortCode)
	statsKey := fmt.Sprintf("stats:%s", shortCode)
	_ = s.cache.Delete(statsKey)

	return nil
}

// GetTopURLs retrieves the most accessed URLs
func (s *URLShortener) GetTopURLs(limit int) ([]models.URLInfo, error) {
	urls, err := s.db.GetTopURLs(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top URLs: %w", err)
	}

	// Add short URL to each entry
	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", s.baseURL, urls[i].ShortCode)
	}

	return urls, nil
}

// GetServiceStats retrieves overall service statistics
func (s *URLShortener) GetServiceStats() (*models.ServiceStats, error) {
	dbStats, err := s.db.GetDatabaseStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	cacheStats, err := s.cache.GetCacheStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	return &models.ServiceStats{
		Database: *dbStats,
		Cache:    *cacheStats,
	}, nil
}
