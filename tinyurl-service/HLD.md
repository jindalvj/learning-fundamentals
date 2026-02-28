# TinyURL Service - High-Level Design (HLD)

## 1. System Overview

A URL shortening service that converts long URLs into short, memorable URLs and redirects users when they access the short URL.

### Key Requirements

**Functional Requirements:**
- Generate a unique short URL for any given long URL
- Redirect users from short URL to original long URL
- Optional: Custom aliases, URL expiration, analytics

**Non-Functional Requirements:**
- High availability (99.9% uptime)
- Low latency (<100ms for redirects)
- Scalable (handle millions of URLs)
- Durable (no data loss)

## 2. Capacity Estimation

**Assumptions:**
- 100M new URLs per month
- Read:Write ratio = 100:1
- URL storage: ~500 bytes per entry
- Data retention: 5 years

**Calculations:**
- Write QPS: 100M / (30 * 24 * 3600) ≈ 40 URLs/sec
- Read QPS: 40 * 100 = 4000 redirects/sec
- Storage (5 years): 100M * 12 * 5 * 500 bytes ≈ 3 TB
- Bandwidth: 4000 req/sec * 500 bytes ≈ 2 MB/sec

## 3. High-Level Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  Load Balancer  │
└──────┬──────────┘
       │
       ▼
┌─────────────────────────┐
│   API Servers (Fleet)   │
│  - URL Creation API     │
│  - URL Redirect API     │
└──────┬──────────────────┘
       │
       ├─────────────┬──────────────┐
       ▼             ▼              ▼
┌────────────┐ ┌──────────┐ ┌─────────────┐
│   Cache    │ │ Database │ │ ID Generator│
│  (Redis)   │ │(Postgres)│ │  Service    │
└────────────┘ └──────────┘ └─────────────┘
```

## 4. Core Components

### 4.1 API Gateway / Load Balancer
- Distributes traffic across multiple API servers
- SSL termination
- Rate limiting
- Request routing

### 4.2 API Servers
- **POST /api/shorten**: Create short URL
- **GET /{shortCode}**: Redirect to original URL
- **GET /api/stats/{shortCode}**: Get analytics (optional)
- Stateless design for horizontal scaling

### 4.3 Database (PostgreSQL)
- Primary data store for URL mappings
- ACID compliance ensures data integrity
- Supports complex queries for analytics
- Replication for high availability

### 4.4 Cache Layer (Redis)
- Cache frequently accessed URLs
- Reduces database load
- TTL-based eviction
- Cache-aside pattern

### 4.5 ID Generator Service
- Generates unique short codes
- Options: Base62 encoding, MD5 hash, Counter-based
- Ensures no collisions

## 5. Short Code Generation Strategies

### Option 1: Base62 Encoding (Recommended)
- Use auto-increment ID from database
- Encode to Base62 (a-z, A-Z, 0-9)
- 7 characters = 62^7 ≈ 3.5 trillion combinations
- Predictable, sequential, no collisions

### Option 2: Hash-based (MD5/SHA)
- Hash the original URL
- Take first 7 characters
- Handle collisions with additional logic
- Non-predictable

### Option 3: Random Generation
- Generate random 7-character string
- Check for collisions
- Retry if collision occurs

## 6. Database Schema Design

### Primary Table: urls
```sql
id (BIGSERIAL)      - Primary key, auto-increment
short_code (VARCHAR(10)) - Unique index, the short identifier
original_url (TEXT) - The long URL
user_id (BIGINT)    - Optional, for authenticated users
created_at (TIMESTAMP)
expires_at (TIMESTAMP) - Optional expiration
is_active (BOOLEAN) - Soft delete flag
```

### Analytics Table: url_stats (Optional)
```sql
id (BIGSERIAL)
short_code (VARCHAR(10))
access_count (BIGINT)
last_accessed (TIMESTAMP)
```

### Click Events Table: clicks (Optional)
```sql
id (BIGSERIAL)
short_code (VARCHAR(10))
clicked_at (TIMESTAMP)
ip_address (VARCHAR(45))
user_agent (TEXT)
referrer (TEXT)
```

## 7. API Design

### Create Short URL
```
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_alias": "mylink" (optional),
  "expires_in_days": 30 (optional)
}

Response:
{
  "short_url": "http://tiny.url/abc1234",
  "short_code": "abc1234",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2025-02-15T10:00:00Z"
}
```

### Redirect
```
GET /{shortCode}

Response: 301/302 Redirect to original URL
Location: https://example.com/very/long/url
```

## 8. Scalability Considerations

### Read-Heavy Optimization
- **Caching**: Redis cache for hot URLs (80-20 rule)
- **CDN**: Serve redirects from edge locations
- **Database Read Replicas**: Multiple read replicas for query distribution

### Write Optimization
- **Asynchronous Processing**: Queue URL creation for batch processing
- **Database Sharding**: Shard by short_code hash
- **Connection Pooling**: Efficient database connection management

### High Availability
- **Multi-AZ Deployment**: Deploy across availability zones
- **Database Replication**: Master-slave replication
- **Health Checks**: Automatic failover

## 9. Data Flow

### URL Creation Flow
1. Client sends POST request with long URL
2. Load balancer routes to available API server
3. API server validates URL
4. Generate unique short code (via DB sequence or ID service)
5. Store mapping in database
6. Return short URL to client
7. (Optional) Pre-warm cache with new entry

### URL Redirect Flow
1. Client requests GET /{shortCode}
2. Check Redis cache
3. If cache hit: Return cached URL with 301 redirect
4. If cache miss:
   - Query database for short_code
   - Store in cache with TTL
   - Return URL with 301 redirect
5. (Optional) Async: Increment access counter

## 10. Technology Stack

- **Language**: Python (Flask/FastAPI) or Node.js (Express)
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **Load Balancer**: Nginx / AWS ALB
- **Container**: Docker
- **Orchestration**: Kubernetes (optional)

## 11. Monitoring & Observability

- **Metrics**: Request rate, latency, error rate, cache hit ratio
- **Logging**: Centralized logging (ELK stack)
- **Alerting**: Alert on high error rates, database connection issues
- **Tracing**: Distributed tracing for debugging

## 12. Security Considerations

- **Rate Limiting**: Prevent abuse (per IP, per user)
- **URL Validation**: Prevent malicious URLs, phishing
- **DDoS Protection**: Cloudflare, AWS Shield
- **HTTPS**: Enforce SSL/TLS
- **Input Sanitization**: Prevent SQL injection, XSS
