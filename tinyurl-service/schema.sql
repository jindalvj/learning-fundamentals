-- TinyURL Database Schema
-- PostgreSQL 14+

-- Drop tables if they exist (for development)
DROP TABLE IF EXISTS clicks CASCADE;
DROP TABLE IF EXISTS url_stats CASCADE;
DROP TABLE IF EXISTS urls CASCADE;

-- Main URLs table
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    original_url TEXT NOT NULL,
    user_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Constraints
    CONSTRAINT check_url_length CHECK (LENGTH(original_url) > 0 AND LENGTH(original_url) <= 2048),
    CONSTRAINT check_short_code_length CHECK (LENGTH(short_code) >= 4 AND LENGTH(short_code) <= 10)
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_short_code ON urls(short_code) WHERE is_active = TRUE;
CREATE INDEX idx_user_id ON urls(user_id);
CREATE INDEX idx_created_at ON urls(created_at DESC);
CREATE INDEX idx_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_original_url_hash ON urls(md5(original_url));

-- URL statistics table
CREATE TABLE url_stats (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    access_count BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMP,
    
    -- Foreign key
    CONSTRAINT fk_short_code FOREIGN KEY (short_code) 
        REFERENCES urls(short_code) ON DELETE CASCADE,
    CONSTRAINT unique_short_code UNIQUE (short_code)
);

CREATE INDEX idx_stats_short_code ON url_stats(short_code);
CREATE INDEX idx_stats_access_count ON url_stats(access_count DESC);

-- Detailed click tracking table (optional - for analytics)
CREATE TABLE clicks (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    clicked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    country_code VARCHAR(2),
    
    -- Foreign key
    CONSTRAINT fk_clicks_short_code FOREIGN KEY (short_code) 
        REFERENCES urls(short_code) ON DELETE CASCADE
);

CREATE INDEX idx_clicks_short_code ON clicks(short_code);
CREATE INDEX idx_clicks_timestamp ON clicks(clicked_at DESC);

-- Function to increment access count
CREATE OR REPLACE FUNCTION increment_url_stats(p_short_code VARCHAR)
RETURNS VOID AS $$
BEGIN
    INSERT INTO url_stats (short_code, access_count, last_accessed)
    VALUES (p_short_code, 1, NOW())
    ON CONFLICT (short_code) 
    DO UPDATE SET 
        access_count = url_stats.access_count + 1,
        last_accessed = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to clean up expired URLs (run via cron job)
CREATE OR REPLACE FUNCTION cleanup_expired_urls()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    UPDATE urls 
    SET is_active = FALSE
    WHERE expires_at IS NOT NULL 
      AND expires_at < NOW() 
      AND is_active = TRUE;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Insert some sample data for testing
INSERT INTO urls (short_code, original_url, created_at) VALUES
    ('test123', 'https://www.example.com/very/long/url/that/needs/shortening', NOW()),
    ('github', 'https://github.com', NOW()),
    ('wiki', 'https://en.wikipedia.org/wiki/URL_shortening', NOW());

-- Initialize stats for sample data
INSERT INTO url_stats (short_code, access_count, last_accessed) VALUES
    ('test123', 0, NOW()),
    ('github', 0, NOW()),
    ('wiki', 0, NOW());

-- Show table statistics
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
