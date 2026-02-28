package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/tinyurl/models"
	"github.com/yourusername/tinyurl/service"
)

// Handler contains all API handlers
type Handler struct {
	service *service.URLShortener
}

// NewHandler creates a new API handler
func NewHandler(service *service.URLShortener) *Handler {
	return &Handler{
		service: service,
	}
}

// SetupRouter configures all routes
func (h *Handler) SetupRouter() *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(corsMiddleware())

	// Routes
	router.GET("/", h.home)
	router.GET("/health", h.healthCheck)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/shorten", h.createShortURL)
		api.GET("/stats/:shortCode", h.getStats)
		api.GET("/info/:shortCode", h.getURLInfo)
		api.DELETE("/url/:shortCode", h.deleteURL)
		api.GET("/top", h.getTopURLs)
		api.GET("/service-stats", h.getServiceStats)
	}

	// Redirect route
	router.GET("/:shortCode", h.redirectToURL)

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// home displays API documentation
func (h *Handler) home(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>TinyURL Service</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-left: 4px solid #007bff; }
        .method { color: #007bff; font-weight: bold; }
        code { background: #eee; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🔗 TinyURL Service API (Go)</h1>
    <p>URL shortening service with caching and analytics.</p>
    
    <h2>API Endpoints</h2>
    
    <div class="endpoint">
        <p><span class="method">POST</span> <code>/api/shorten</code></p>
        <p>Create a short URL</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">GET</span> <code>/{shortCode}</code></p>
        <p>Redirect to original URL</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">GET</span> <code>/api/stats/{shortCode}</code></p>
        <p>Get statistics for a short URL</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">GET</span> <code>/api/info/{shortCode}</code></p>
        <p>Get complete information about a short URL</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">DELETE</span> <code>/api/url/{shortCode}</code></p>
        <p>Delete a short URL</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">GET</span> <code>/api/top</code></p>
        <p>Get top accessed URLs</p>
    </div>
    
    <div class="endpoint">
        <p><span class="method">GET</span> <code>/api/service-stats</code></p>
        <p>Get overall service statistics</p>
    </div>
    
    <h2>Status</h2>
    <p>Service is running ✓</p>
</body>
</html>
    `
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// healthCheck checks service health
func (h *Handler) healthCheck(c *gin.Context) {
	// In a real implementation, check database and cache connectivity
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
		"cache":    "connected",
	})
}

// createShortURL handles POST /api/shorten
func (h *Handler) createShortURL(c *gin.Context) {
	var req models.CreateURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Create short URL
	result, err := h.service.CreateShortURL(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    result,
	}

	if result.IsExisting {
		response.Message = "Existing short URL returned"
	}

	c.JSON(http.StatusCreated, response)
}

// redirectToURL handles GET /:shortCode
func (h *Handler) redirectToURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	// Get client information for analytics
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()
	referrer := c.Request.Referer()

	var ip, ua, ref *string
	if ipAddress != "" {
		ip = &ipAddress
	}
	if userAgent != "" {
		ua = &userAgent
	}
	if referrer != "" {
		ref = &referrer
	}

	// Get original URL
	originalURL, err := h.service.GetOriginalURL(shortCode, true, ip, ua, ref)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Short URL not found or has expired",
		})
		return
	}

	// Redirect (301 for permanent, 302 for temporary)
	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// getStats handles GET /api/stats/:shortCode
func (h *Handler) getStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	stats, err := h.service.GetStats(shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if stats == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Statistics not found for this short code",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// getURLInfo handles GET /api/info/:shortCode
func (h *Handler) getURLInfo(c *gin.Context) {
	shortCode := c.Param("shortCode")

	info, err := h.service.GetURLInfo(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Short URL not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    info,
	})
}

// deleteURL handles DELETE /api/url/:shortCode
func (h *Handler) deleteURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	err := h.service.DeleteURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Short URL not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Short URL deleted successfully",
	})
}

// getTopURLs handles GET /api/top
func (h *Handler) getTopURLs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	urls, err := h.service.GetTopURLs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    urls,
	})
}

// getServiceStats handles GET /api/service-stats
func (h *Handler) getServiceStats(c *gin.Context) {
	stats, err := h.service.GetServiceStats()
	if err != nil {
		log.Printf("Error getting service stats: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve service statistics",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}
