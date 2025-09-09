package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Host string
	Port string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type AppConfig struct {
	BaseURL                string
	DefaultExpirationMin   int
	ShortCodeLength        int
	Environment            string
	CleanupIntervalMinutes int
}

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	App      AppConfig
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return defaultValue
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environments variables.")
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_DBNAME", "url_shortener"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		App: AppConfig{
			BaseURL:                getEnv("BASE_URL", "http://localhost:8080"),
			DefaultExpirationMin:   getEnvAsInt("DEFAULT_EXPIRATION_MINUTES", 1),
			ShortCodeLength:        getEnvAsInt("SHORT_CODE_LENGTH", 8),
			Environment:            getEnv("ENV", "development"),
			CleanupIntervalMinutes: getEnvAsInt("CLEANUP_INTERVAL_MINUTES", 1),
		},
	}

	return config
}
