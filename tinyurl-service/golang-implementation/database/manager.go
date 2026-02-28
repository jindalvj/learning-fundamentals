package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/tinyurl/models"
)

// Manager handles all database operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new database manager with connection pooling
func NewManager(databaseURL string) (*Manager, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Manager{db: db}, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// InsertURL inserts a new URL and returns the auto-generated ID
func (m *Manager) InsertURL(originalURL string, shortCode string, userID *int64, expiresInDays *int) (int64, error) {
	var expiresAt *time.Time
	if expiresInDays != nil {
		expires := time.Now().AddDate(0, 0, *expiresInDays)
		expiresAt = &expires
	}

	query := `
		INSERT INTO urls (short_code, original_url, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := m.db.QueryRow(query, shortCode, originalURL, userID, expiresAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert URL: %w", err)
	}

	return id, nil
}

// UpdateShortCode updates the short_code for a given URL ID
func (m *Manager) UpdateShortCode(urlID int64, shortCode string) error {
	query := `UPDATE urls SET short_code = $1 WHERE id = $2`
	
	result, err := m.db.Exec(query, shortCode, urlID)
	if err != nil {
		return fmt.Errorf("failed to update short code: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}

// GetURLByShortCode retrieves URL data by short code
func (m *Manager) GetURLByShortCode(shortCode string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active
		FROM urls
		WHERE short_code = $1
	`

	url := &models.URL{}
	err := m.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.UserID,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return url, nil
}

// GetURLByOriginalURL finds existing short code for a given original URL
func (m *Manager) GetURLByOriginalURL(originalURL string) (*models.URL, error) {
	query := `
		SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active
		FROM urls
		WHERE original_url = $1 
		  AND is_active = TRUE
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
		LIMIT 1
	`

	url := &models.URL{}
	err := m.db.QueryRow(query, originalURL).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.UserID,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL by original: %w", err)
	}

	return url, nil
}

// CheckAliasAvailable checks if a custom alias is available
func (m *Manager) CheckAliasAvailable(alias string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM urls 
			WHERE short_code = $1 AND is_active = TRUE
		)
	`

	var exists bool
	err := m.db.QueryRow(query, alias).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check alias: %w", err)
	}

	return !exists, nil
}

// DeactivateURL soft deletes a URL
func (m *Manager) DeactivateURL(shortCode string) error {
	query := `UPDATE urls SET is_active = FALSE WHERE short_code = $1`
	
	result, err := m.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to deactivate URL: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}

// InitializeStats creates an initial stats entry for a new URL
func (m *Manager) InitializeStats(shortCode string) error {
	query := `
		INSERT INTO url_stats (short_code, access_count, last_accessed)
		VALUES ($1, 0, NOW())
		ON CONFLICT (short_code) DO NOTHING
	`

	_, err := m.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to initialize stats: %w", err)
	}

	return nil
}

// IncrementAccessCount increments the access counter for a URL
func (m *Manager) IncrementAccessCount(shortCode string) error {
	query := `SELECT increment_url_stats($1)`
	
	_, err := m.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment access count: %w", err)
	}

	return nil
}

// GetStats retrieves statistics for a URL
func (m *Manager) GetStats(shortCode string) (*models.URLStats, error) {
	query := `
		SELECT short_code, access_count, last_accessed
		FROM url_stats
		WHERE short_code = $1
	`

	stats := &models.URLStats{}
	err := m.db.QueryRow(query, shortCode).Scan(
		&stats.ShortCode,
		&stats.AccessCount,
		&stats.LastAccessed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

// GetTopURLs retrieves the most accessed URLs
func (m *Manager) GetTopURLs(limit int) ([]models.URLInfo, error) {
	query := `
		SELECT 
			u.short_code,
			u.original_url,
			u.created_at,
			s.access_count,
			s.last_accessed
		FROM urls u
		JOIN url_stats s ON u.short_code = s.short_code
		WHERE u.is_active = TRUE
		ORDER BY s.access_count DESC
		LIMIT $1
	`

	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top URLs: %w", err)
	}
	defer rows.Close()

	var urls []models.URLInfo
	for rows.Next() {
		var info models.URLInfo
		var stats models.URLStats

		err := rows.Scan(
			&info.ShortCode,
			&info.OriginalURL,
			&info.CreatedAt,
			&stats.AccessCount,
			&stats.LastAccessed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		stats.ShortCode = info.ShortCode
		info.Stats = &stats
		urls = append(urls, info)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return urls, nil
}

// LogClick logs a detailed click event
func (m *Manager) LogClick(shortCode string, ipAddress, userAgent, referrer *string) error {
	query := `
		INSERT INTO clicks (short_code, ip_address, user_agent, referrer)
		VALUES ($1, $2, $3, $4)
	`

	_, err := m.db.Exec(query, shortCode, ipAddress, userAgent, referrer)
	if err != nil {
		return fmt.Errorf("failed to log click: %w", err)
	}

	return nil
}

// GetDatabaseStats retrieves overall database statistics
func (m *Manager) GetDatabaseStats() (*models.DatabaseStats, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM urls WHERE is_active = TRUE) as active_urls,
			(SELECT COUNT(*) FROM urls WHERE is_active = FALSE) as inactive_urls,
			(SELECT COALESCE(SUM(access_count), 0) FROM url_stats) as total_clicks,
			(SELECT COUNT(*) FROM clicks) as total_click_events
	`

	stats := &models.DatabaseStats{}
	err := m.db.QueryRow(query).Scan(
		&stats.ActiveURLs,
		&stats.InactiveURLs,
		&stats.TotalClicks,
		&stats.TotalClickEvents,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return stats, nil
}

// CleanupExpiredURLs marks expired URLs as inactive
func (m *Manager) CleanupExpiredURLs() (int64, error) {
	query := `SELECT cleanup_expired_urls()`
	
	var count int64
	err := m.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired URLs: %w", err)
	}

	return count, nil
}
