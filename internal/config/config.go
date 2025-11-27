package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBUrlMigration string
	JWTSecret      string

	DBHost     string
	DBUser     string
	DBName     string
	DBPassword string
	DBPort     string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file")
	}

	log.Println("config loaded")

	return &Config{
		Port:           GetEnv("APP_PORT", "8080"),
		DBUrlMigration: GetEnv("DATABASE_URL", "mysql://root:go_tweets@127.0.0.1:3306/go_tweets?parseTime=true"),
		JWTSecret:      GetEnv("JWT_SECRET", "super_secret_123"),

		DBHost:     GetEnv("DB_HOST", "localhost"),
		DBUser:     GetEnv("DB_USER", "root"),
		DBName:     GetEnv("DB_NAME", "go_tweets"),
		DBPassword: GetEnv("DB_PASSWORD", "go-tweets"),
		DBPort:     GetEnv("DB_PORT", "3306"),
	}, nil
}
