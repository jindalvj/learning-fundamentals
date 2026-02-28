# 🔗 TinyURL Service

A production-ready URL shortening service built with Python, Flask, PostgreSQL, and Redis. This project demonstrates both High-Level Design (HLD) and Low-Level Design (LLD) principles for building scalable systems.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Design Documents](#design-documents)
- [Performance](#performance)
- [Testing](#testing)
- [Deployment](#deployment)

## ✨ Features

- **URL Shortening**: Convert long URLs into short, memorable links
- **Custom Aliases**: Create branded short links with custom aliases
- **URL Expiration**: Set expiration dates for temporary links
- **Analytics**: Track click counts and access patterns
- **High Performance**: Redis caching for sub-10ms response times
- **Scalability**: Horizontal scaling support with stateless design
- **RESTful API**: Clean, well-documented REST API
- **Docker Support**: One-command deployment with Docker Compose

## 🏗️ Architecture

### System Components

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
└──────┬──────────────────┘
       │
       ├─────────────┬──────────────┐
       ▼             ▼              ▼
┌────────────┐ ┌──────────┐ ┌─────────────┐
│   Redis    │ │ Postgres │ │Base62 Encoder│
│   Cache    │ │ Database │ │             │
└────────────┘ └──────────┘ └─────────────┘
```

### Design Patterns

- **Cache-Aside Pattern**: For optimal read performance
- **Base62 Encoding**: For collision-free short code generation
- **Connection Pooling**: For efficient database operations
- **Async Operations**: For non-blocking statistics updates

## 🛠️ Tech Stack

- **Backend**: Python 3.11, Flask
- **Database**: PostgreSQL 14
- **Cache**: Redis 7
- **Containerization**: Docker, Docker Compose
- **Testing**: pytest

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Python 3.11+ (for local development)
- PostgreSQL 14+ (for local development)
- Redis 7+ (for local development)

### Quick Start with Docker

1. **Clone the repository**
```bash
git clone <repository-url>
cd tinyurl-service
```

2. **Start services with Docker Compose**
```bash
docker-compose up -d
```

3. **Verify services are running**
```bash
# Check API health
curl http://localhost:5000/health

# Expected response:
# {
#   "status": "healthy",
#   "database": "connected",
#   "cache": "connected"
# }
```

4. **Create your first short URL**
```bash
curl -X POST http://localhost:5000/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.example.com/very/long/url"}'

# Response:
# {
#   "success": true,
#   "data": {
#     "short_code": "000001",
#     "short_url": "http://localhost:5000/000001",
#     "original_url": "https://www.example.com/very/long/url",
#     "created_at": "2025-02-15T10:00:00"
#   }
# }
```

5. **Test the redirect**
```bash
curl -L http://localhost:5000/000001
# Redirects to: https://www.example.com/very/long/url
```

### Local Development Setup

1. **Create virtual environment**
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. **Install dependencies**
```bash
pip install -r requirements.txt
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your database and Redis credentials
```

4. **Initialize database**
```bash
psql -U postgres -d tinyurl < schema.sql
```

5. **Run the application**
```bash
python api.py
```

The API will be available at `http://localhost:5000`

## 📖 API Documentation

### Create Short URL

**Endpoint:** `POST /api/shorten`

**Request Body:**
```json
{
  "url": "https://example.com/long/url",
  "custom_alias": "mylink",      // Optional
  "expires_in_days": 30,          // Optional
  "user_id": 123                   // Optional
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "short_code": "abc123",
    "short_url": "http://tiny.url/abc123",
    "original_url": "https://example.com/long/url",
    "created_at": "2025-02-15T10:00:00"
  }
}
```

### Redirect to Original URL

**Endpoint:** `GET /{shortCode}`

**Response:** `301 Redirect` to original URL

### Get Statistics

**Endpoint:** `GET /api/stats/{shortCode}`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "short_code": "abc123",
    "access_count": 42,
    "last_accessed": "2025-02-15T10:00:00"
  }
}
```

### Get URL Information

**Endpoint:** `GET /api/info/{shortCode}`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "short_code": "abc123",
    "original_url": "https://example.com",
    "short_url": "http://tiny.url/abc123",
    "created_at": "2025-02-15T10:00:00",
    "is_active": true,
    "stats": {
      "access_count": 42,
      "last_accessed": "2025-02-15T10:00:00"
    }
  }
}
```

### Delete Short URL

**Endpoint:** `DELETE /api/url/{shortCode}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Short URL deleted successfully"
}
```

### Get Top URLs

**Endpoint:** `GET /api/top?limit=10`

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "short_code": "abc123",
      "original_url": "https://example.com",
      "access_count": 100,
      "created_at": "2025-02-15T10:00:00"
    }
  ]
}
```

### Get Service Statistics

**Endpoint:** `GET /api/service-stats`

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "database": {
      "active_urls": 1000,
      "total_clicks": 50000
    },
    "cache": {
      "hit_rate": "85.5%",
      "total_keys": 500
    }
  }
}
```

## 📚 Design Documents

### High-Level Design (HLD)
See [HLD.md](HLD.md) for:
- System architecture and components
- Capacity estimation and scaling strategies
- Data flow diagrams
- Technology choices and trade-offs

### Low-Level Design (LLD)
See [LLD.md](LLD.md) for:
- Detailed class diagrams
- Database schema design
- Algorithm implementations
- API specifications
- Error handling strategies

## ⚡ Performance

### Short Code Generation
- **Base62 Encoding**: Supports 62^7 = 3.5 trillion URLs with 7 characters
- **Time Complexity**: O(log n) for encoding/decoding

### URL Lookup Performance
- **Cache Hit**: < 5ms
- **Cache Miss**: < 50ms (with database query)
- **Expected Cache Hit Rate**: 80-90% for popular URLs

### Scalability
- **Horizontal Scaling**: Stateless API servers
- **Database**: Read replicas for query distribution
- **Cache**: Redis cluster for distributed caching
- **Estimated Capacity**: 
  - 4000 redirects/second per server
  - 40 URL creations/second per server

## 🧪 Testing

### Run Tests
```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=. --cov-report=html

# Run specific test file
pytest test_base62_encoder.py
```

### Test Base62 Encoder
```bash
python base62_encoder.py
```

### Manual Testing
```bash
# Create URL
curl -X POST http://localhost:5000/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'

# Get stats
curl http://localhost:5000/api/stats/000001

# Test redirect
curl -I http://localhost:5000/000001
```

## 🚢 Deployment

### Production Deployment Checklist

- [ ] Set `DEBUG=false` in environment
- [ ] Use strong database credentials
- [ ] Enable SSL/TLS (HTTPS)
- [ ] Set up monitoring and alerting
- [ ] Configure proper rate limiting
- [ ] Set up database backups
- [ ] Configure Redis persistence
- [ ] Use a reverse proxy (Nginx)
- [ ] Set up health checks
- [ ] Configure auto-scaling

### Docker Production Deployment

```bash
# Build production image
docker build -t tinyurl-api:latest .

# Run with production configuration
docker-compose -f docker-compose.prod.yml up -d
```

### Environment Variables

See [.env.example](.env.example) for all configuration options.

## 📊 Database Schema

### Main Tables

**urls** - Stores URL mappings
- `id` (BIGSERIAL): Primary key
- `short_code` (VARCHAR): Unique short identifier
- `original_url` (TEXT): Long URL
- `created_at` (TIMESTAMP): Creation time
- `expires_at` (TIMESTAMP): Expiration time
- `is_active` (BOOLEAN): Soft delete flag

**url_stats** - Stores access statistics
- `id` (BIGSERIAL): Primary key
- `short_code` (VARCHAR): Reference to urls
- `access_count` (BIGINT): Click counter
- `last_accessed` (TIMESTAMP): Last access time

**clicks** - Detailed analytics (optional)
- `id` (BIGSERIAL): Primary key
- `short_code` (VARCHAR): Reference to urls
- `clicked_at` (TIMESTAMP): Click timestamp
- `ip_address` (VARCHAR): Visitor IP
- `user_agent` (TEXT): Browser info
- `referrer` (TEXT): Referrer URL

## 🔧 Troubleshooting

### Common Issues

**Database connection failed**
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres
```

**Redis connection failed**
```bash
# Check if Redis is running
docker-compose ps redis

# Test Redis
docker-compose exec redis redis-cli ping
```

**Port already in use**
```bash
# Change port in docker-compose.yml or .env
# Default port: 5000
```

## 📈 Monitoring

### Metrics to Track
- Request rate (requests/second)
- Response time (p50, p95, p99)
- Cache hit rate
- Database connection pool usage
- Error rate

### Logging
Logs are written to stdout and can be viewed with:
```bash
docker-compose logs -f api
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📝 License

This project is open source and available under the MIT License.

## 🎓 Learning Resources

This project demonstrates:
- System Design principles
- Database design and indexing
- Caching strategies
- RESTful API design
- Docker containerization
- Horizontal scaling
- Performance optimization

Perfect for:
- Learning system design
- Interview preparation
- Understanding distributed systems
- Building production services

## 📞 Support

For questions or issues, please open an issue on GitHub.

---

Built with ❤️ for learning system design
