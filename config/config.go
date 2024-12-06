package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ZenrowsApiKey   string
	MongoURL        string
	MongoDBName     string
	MongoCollection string
}

var (
	config *Config
	once   sync.Once
)

func Get() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}

		config = &Config{
			ZenrowsApiKey:   getEnvOrFatal("ZENROWS_API_KEY"),
			MongoURL:        getEnvOrFatal("MONGODB_URL"),
			MongoDBName:     getEnvOrFatal("MONGODB_DB"),
			MongoCollection: getEnvOrFatal("MONGODB_COLLECTION"),
		}
	})
	return config
}

func getEnvOrFatal(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s not found in environment", key)
	}
	return value
}
