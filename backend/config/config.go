package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
