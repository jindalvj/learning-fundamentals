package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourusername/tinyurl/api"
	"github.com/yourusername/tinyurl/cache"
	"github.com/yourusername/tinyurl/database"
	"github.com/yourusername/tinyurl/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration from environment
	databaseURL := getEnv("DATABASE_URL", "postgresql://tinyurl_user:tinyurl_password@localhost:5432/tinyurl?sslmode=disable")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnvAsInt("REDIS_PORT", 6379)
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvAsInt("REDIS_DB", 0)
	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	port := getEnv("PORT", "8080")

	// Initialize database
	log.Println("Connecting to database...")
	db, err := database.NewManager(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Database connected")

	// Initialize cache
	log.Println("Connecting to Redis...")
	cacheManager, err := cache.NewManager(redisHost, redisPort, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cacheManager.Close()
	log.Println("✓ Redis connected")

	// Initialize URL shortener service
	serviceConfig := service.Config{
		BaseURL:             baseURL,
		CacheTTL:            1 * time.Hour,
		StatsCacheTTL:       5 * time.Minute,
		EnableDeduplication: true,
		MinShortCodeLength:  6,
	}
	urlService := service.NewURLShortener(db, cacheManager, serviceConfig)
	log.Println("✓ URL Shortener service initialized")

	// Setup API handlers
	handler := api.NewHandler(urlService)
	router := handler.SetupRouter()

	// Print startup banner
	printBanner(port, baseURL)

	// Start server
	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// printBanner prints a startup banner
func printBanner(port, baseURL string) {
	banner := fmt.Sprintf(`
╔════════════════════════════════════════╗
║     TinyURL Service API Started       ║
╠════════════════════════════════════════╣
║  URL: http://localhost:%s           ║
║  Base URL: %-27s ║
║  Environment: Production              ║
╚════════════════════════════════════════╝
`, port, baseURL)
	fmt.Println(banner)
}
