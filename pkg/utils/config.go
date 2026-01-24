package utils

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	DatabaseURL string
	OllamaURL   string
}

func LoadConfig() (*Config, error) {
	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5433"),
		DBUser:     getEnv("DB_USER", "enron"),
		DBPassword: getEnv("DB_PASSWORD", "enron123"),
		DBName:     getEnv("DB_NAME", "enron_graph"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		OllamaURL:  getEnv("OLLAMA_URL", "http://localhost:11434"),
	}

	// Build DatabaseURL
	config.DatabaseURL = config.PostgresURL()

	return config, nil
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) PostgresURL() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
