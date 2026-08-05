package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values loaded from environment variables
type Config struct {
	// Database
	DatabaseURL string

	// JWT
	JWTSecret string

	// Admin seed
	AdminUsername string
	AdminEmail    string
	AdminPassword string
	AdminFullName string

	// Server
	Port    string
	GinMode string

	// CORS
	CORSOrigins []string

	// Sandbox
	SandboxTimeout  int
	SandboxMemoryMB int
	SandboxWorkers  int
}

// AppConfig is the global configuration instance
var AppConfig Config

// Load reads environment variables and populates AppConfig.
// Call this at the very start of main().
func Load() {
	// Load .env file if it exists (does not override existing env vars)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	AppConfig = Config{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://atom_user:atom_password@localhost:5432/atom_db?sslmode=disable"),

		// JWT
		JWTSecret: getEnvRequired("JWT_SECRET"),

		// Admin seed
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "Admin123!"),
		AdminFullName: getEnv("ADMIN_FULLNAME", "Admin"),

		// Server
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		// CORS
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:5173"), ","),

		// Sandbox
		SandboxTimeout:  getEnvInt("SANDBOX_TIMEOUT", 30),
		SandboxMemoryMB: getEnvInt("SANDBOX_MEMORY_MB", 256),
		SandboxWorkers:  getEnvInt("SANDBOX_WORKERS", 3),
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: Required environment variable %s is not set. Check your .env file.", key)
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if num, err := strconv.Atoi(val); err == nil {
			return num
		}
	}
	return fallback
}
