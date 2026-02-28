# TinyURL Service - Low-Level Design (LLD)

## 1. Class Diagram

```
┌─────────────────────┐
│   URLShortener      │
├─────────────────────┤
│ - db_connection     │
│ - cache_connection  │
├─────────────────────┤
│ + create_short_url()│
│ + get_original_url()│
│ + encode_id()       │
│ + decode_id()       │
└─────────────────────┘
         │
         │ uses
         ▼
┌─────────────────────┐
│   DatabaseManager   │
├─────────────────────┤
│ - connection_pool   │
├─────────────────────┤
│ + insert_url()      │
│ + get_url()         │
│ + update_stats()    │
└─────────────────────┘

┌─────────────────────┐
│   CacheManager      │
├─────────────────────┤
│ - redis_client      │
├─────────────────────┤
│ + get()             │
│ + set()             │
│ + delete()          │
└─────────────────────┘

┌─────────────────────┐
│   Base62Encoder     │
├─────────────────────┤
│ - BASE62_CHARS      │
├─────────────────────┤
│ + encode()          │
│ + decode()          │
└─────────────────────┘
```

## 2. Database Schema (Detailed)

### Table: urls

```sql
CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    user_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    CONSTRAINT check_url_length CHECK (LENGTH(original_url) > 0),
    CONSTRAINT check_short_code_length CHECK (LENGTH(short_code) >= 4)
);

-- Indexes
CREATE UNIQUE INDEX idx_short_code ON urls(short_code) WHERE is_active = TRUE;
CREATE INDEX idx_user_id ON urls(user_id);
CREATE INDEX idx_created_at ON urls(created_at);
CREATE INDEX idx_original_url_hash ON urls(md5(original_url));
```

### Table: url_stats

```sql
CREATE TABLE url_stats (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    access_count BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMP,
    
    FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE,
    CONSTRAINT unique_short_code UNIQUE (short_code)
);

CREATE INDEX idx_stats_short_code ON url_stats(short_code);
```

### Table: clicks (Optional - for detailed analytics)

```sql
CREATE TABLE clicks (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    clicked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    country_code VARCHAR(2),
    
    FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE
);

CREATE INDEX idx_clicks_short_code ON clicks(short_code);
CREATE INDEX idx_clicks_timestamp ON clicks(clicked_at);

-- Partition by month for better performance
CREATE TABLE clicks_2025_02 PARTITION OF clicks
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
```

## 3. Core Algorithms

### 3.1 Base62 Encoding Algorithm

```python
class Base62Encoder:
    """
    Converts numeric IDs to Base62 strings and vice versa.
    Base62: [a-z, A-Z, 0-9] = 62 characters
    """
    
    BASE62_CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    BASE = 62
    
    @staticmethod
    def encode(num: int) -> str:
        """
        Convert integer to base62 string.
        Example: 125 -> "23" in base62
        
        Time Complexity: O(log(num))
        Space Complexity: O(log(num))
        """
        if num == 0:
            return Base62Encoder.BASE62_CHARS[0]
        
        result = []
        while num > 0:
            remainder = num % Base62Encoder.BASE
            result.append(Base62Encoder.BASE62_CHARS[remainder])
            num //= Base62Encoder.BASE
        
        return ''.join(reversed(result))
    
    @staticmethod
    def decode(short_code: str) -> int:
        """
        Convert base62 string back to integer.
        Example: "23" -> 125
        
        Time Complexity: O(len(short_code))
        Space Complexity: O(1)
        """
        num = 0
        for char in short_code:
            num = num * Base62Encoder.BASE + Base62Encoder.BASE62_CHARS.index(char)
        return num
```

### 3.2 URL Creation Algorithm

```
Algorithm: CREATE_SHORT_URL(original_url, custom_alias=None)

Input: 
    - original_url: The long URL to shorten
    - custom_alias: Optional custom short code

Output: short_code

Steps:
1. Validate original_url
   - Check if URL is valid format
   - Check if URL is reachable (optional)
   - Sanitize for security

2. If custom_alias provided:
   a. Check if alias is available in database
   b. If available, use it as short_code
   c. If not available, return error

3. If no custom_alias:
   a. Check if URL already exists in database (optional deduplication)
   b. If exists and not expired, return existing short_code
   c. If not exists or expired:
      i. Insert new row in database
      ii. Get auto-generated ID
      iii. Encode ID using Base62
      iv. Update row with short_code

4. Initialize stats entry with access_count = 0

5. Return short_code

Time Complexity: O(1) average
Space Complexity: O(1)
```

### 3.3 URL Redirect Algorithm

```
Algorithm: GET_ORIGINAL_URL(short_code)

Input: short_code
Output: original_url or None

Steps:
1. Check Redis cache
   - Key: "url:" + short_code
   - If found and not expired:
     a. Increment access counter (async)
     b. Return cached original_url

2. If cache miss:
   a. Query database for short_code
   b. If not found or expired or not active:
      - Return None (404)
   c. If found:
      - Store in cache with TTL (e.g., 1 hour)
      - Increment access counter (async)
      - Return original_url

3. Log access event (async, optional)

Time Complexity: O(1) average (cache hit), O(log n) worst case (DB query)
Space Complexity: O(1)
```

## 4. Detailed Component Design

### 4.1 URLShortener Service Class

```python
class URLShortener:
    def __init__(self, db_manager, cache_manager, encoder):
        self.db = db_manager
        self.cache = cache_manager
        self.encoder = encoder
        self.cache_ttl = 3600  # 1 hour
    
    def create_short_url(self, original_url: str, 
                        custom_alias: str = None,
                        user_id: int = None,
                        expires_in_days: int = None) -> dict:
        """
        Creates a shortened URL.
        
        Returns:
            {
                'short_code': str,
                'original_url': str,
                'short_url': str,
                'created_at': datetime
            }
        """
        # Validation
        if not self._is_valid_url(original_url):
            raise ValueError("Invalid URL format")
        
        # Handle custom alias
        if custom_alias:
            if not self._is_alias_available(custom_alias):
                raise ValueError("Custom alias already taken")
            short_code = custom_alias
            url_id = None
        else:
            # Check for existing URL (deduplication)
            existing = self.db.get_by_original_url(original_url)
            if existing and existing['is_active']:
                return existing
            
            # Insert and get ID
            url_id = self.db.insert_url(
                original_url=original_url,
                user_id=user_id,
                expires_in_days=expires_in_days
            )
            short_code = self.encoder.encode(url_id)
        
        # Update short_code in database
        self.db.update_short_code(url_id or short_code, short_code)
        
        # Initialize stats
        self.db.initialize_stats(short_code)
        
        return {
            'short_code': short_code,
            'original_url': original_url,
            'short_url': f"http://tiny.url/{short_code}",
            'created_at': datetime.now()
        }
    
    def get_original_url(self, short_code: str) -> str:
        """
        Retrieves original URL from short code.
        Uses cache-aside pattern.
        """
        # Check cache first
        cache_key = f"url:{short_code}"
        cached_url = self.cache.get(cache_key)
        
        if cached_url:
            # Async increment
            self._increment_stats_async(short_code)
            return cached_url
        
        # Cache miss - query database
        url_data = self.db.get_url(short_code)
        
        if not url_data or not url_data['is_active']:
            return None
        
        if url_data['expires_at'] and url_data['expires_at'] < datetime.now():
            return None
        
        original_url = url_data['original_url']
        
        # Update cache
        self.cache.set(cache_key, original_url, ttl=self.cache_ttl)
        
        # Async increment
        self._increment_stats_async(short_code)
        
        return original_url
```

### 4.2 Database Manager

```python
class DatabaseManager:
    def __init__(self, connection_string):
        self.pool = self._create_connection_pool(connection_string)
    
    def insert_url(self, original_url: str, user_id: int = None, 
                   expires_in_days: int = None) -> int:
        """Inserts URL and returns auto-generated ID"""
        with self.pool.get_connection() as conn:
            cursor = conn.cursor()
            
            expires_at = None
            if expires_in_days:
                expires_at = datetime.now() + timedelta(days=expires_in_days)
            
            query = """
                INSERT INTO urls (original_url, user_id, expires_at)
                VALUES (%s, %s, %s)
                RETURNING id
            """
            cursor.execute(query, (original_url, user_id, expires_at))
            url_id = cursor.fetchone()[0]
            conn.commit()
            return url_id
    
    def update_short_code(self, url_id: int, short_code: str):
        """Updates the short_code for a given URL ID"""
        with self.pool.get_connection() as conn:
            cursor = conn.cursor()
            query = "UPDATE urls SET short_code = %s WHERE id = %s"
            cursor.execute(query, (short_code, url_id))
            conn.commit()
    
    def get_url(self, short_code: str) -> dict:
        """Retrieves URL data by short_code"""
        with self.pool.get_connection() as conn:
            cursor = conn.cursor()
            query = """
                SELECT id, short_code, original_url, is_active, 
                       expires_at, created_at
                FROM urls
                WHERE short_code = %s
            """
            cursor.execute(query, (short_code,))
            result = cursor.fetchone()
            
            if not result:
                return None
            
            return {
                'id': result[0],
                'short_code': result[1],
                'original_url': result[2],
                'is_active': result[3],
                'expires_at': result[4],
                'created_at': result[5]
            }
```

## 5. API Endpoint Implementation Pseudocode

### POST /api/shorten

```
ENDPOINT: POST /api/shorten

Request Validation:
- Check if 'url' field exists
- Validate URL format
- Check custom_alias length (if provided)
- Rate limit check

Processing:
1. Extract parameters from request body
2. Call url_shortener.create_short_url()
3. Handle exceptions (alias taken, invalid URL, etc.)

Response:
- 201 Created: Return short URL details
- 400 Bad Request: Invalid input
- 409 Conflict: Custom alias already exists
- 429 Too Many Requests: Rate limit exceeded
```

### GET /{shortCode}

```
ENDPOINT: GET /{shortCode}

Request Validation:
- Validate short_code format

Processing:
1. Call url_shortener.get_original_url(short_code)
2. If URL found:
   - Return 301/302 redirect
3. If URL not found:
   - Return 404 Not Found

Response:
- 301/302 Redirect: With Location header
- 404 Not Found: Short code doesn't exist
- 410 Gone: URL has expired
```

## 6. Concurrency Handling

### Race Condition: Duplicate Short Codes
- **Problem**: Two requests might generate the same short_code
- **Solution**: Database unique constraint on short_code
- **Handling**: Retry with new ID if constraint violation

### High Write Concurrency
- **Solution**: Connection pooling (e.g., 100 connections)
- **Solution**: Database write optimization (batch inserts)

### Cache Stampede
- **Problem**: Many requests for expired cache key hit database
- **Solution**: Lock-based cache refresh
- **Solution**: Probabilistic early expiration

## 7. Error Handling

```python
class URLShortenerException(Exception):
    pass

class InvalidURLException(URLShortenerException):
    pass

class AliasAlreadyExistsException(URLShortenerException):
    pass

class ShortCodeNotFoundException(URLShortenerException):
    pass

class URLExpiredException(URLShortenerException):
    pass
```

## 8. Performance Optimization

### Database Optimization
- Indexes on short_code, created_at, user_id
- Connection pooling (pgbouncer)
- Read replicas for analytics queries
- Partitioning clicks table by date

### Cache Optimization
- Cache hot URLs (80-20 rule)
- TTL: 1 hour for URLs, 5 min for stats
- Cache warming for popular URLs
- LRU eviction policy

### Application Optimization
- Async operations for stats updates
- Batch database operations
- Response compression
- Keep-alive connections

## 9. Testing Strategy

### Unit Tests
- Base62 encoding/decoding
- URL validation
- Database operations
- Cache operations

### Integration Tests
- End-to-end URL creation flow
- Redirect flow with cache
- Expiration handling
- Custom alias handling

### Load Tests
- 5000 req/sec redirect performance
- Cache hit rate measurement
- Database connection pool under load

### Edge Cases
- Expired URLs
- Deleted URLs
- Invalid short codes
- Extremely long URLs
- Special characters in URLs
