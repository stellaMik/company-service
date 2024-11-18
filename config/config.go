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

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	return &Config{
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		KafkaURL:     os.Getenv("KAFKA_URL"),
		APIPort:      os.Getenv("API_PORT"),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		KafkaGroupId: os.Getenv("KAFKA_GROUP_ID"),
		User:         os.Getenv("API_USER"),
		Password:     os.Getenv("API_PASSWORD"),
	}, nil
}
