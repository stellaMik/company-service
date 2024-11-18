package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser       string
	DBPassword   string
	DBName       string
	DBHost       string
	DBPort       string
	JWTSecret    string
	KafkaURL     string
	APIPort      string
	KafkaGroupId string
	KafkaTopic   string
	User         string
	Password     string
}

// LoadConfig loads the configuration from the environment variables
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Return a new Config struct with defaults set if env variables are missing
	return &Config{
		// Database configuration
		DBUser:     getEnv("DB_USER", "user1"),
		DBPassword: getEnv("DB_PASSWORD", "test1"),
		DBName:     getEnv("DB_NAME", "Companies"),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),

		// API Configuration
		APIPort:  getEnv("API_PORT", "8080"),
		User:     getEnv("API_USER", "user2"),
		Password: getEnv("API_PASSWORD", "test2"),

		// JWT Configuration
		JWTSecret: getEnv("JWT_SECRET", "secretTest"),

		// Kafka Configuration
		KafkaURL:     getEnv("KAFKA_URL", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "company_events"),
		KafkaGroupId: getEnv("KAFKA_GROUP_ID", "company_events_group"),
	}, nil
}

// getEnv checks if the environment variable is set, if not returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
