package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseUser     string
	DatabasePassword string
	DatabaseTable    string
	DatabaseAddress  string
	Port             string
}

// Get returns a config object that is built from environment variables
func Get() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	c := &Config{
		DatabaseUser:     os.Getenv("DB_USER"),
		DatabasePassword: os.Getenv("DB_PASS"),
		DatabaseTable:    os.Getenv("DB_TABLE"),
		Port:             os.Getenv("PORT"),
	}

	return c, nil
}
