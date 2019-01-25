package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseUser         string
	DatabaseRootUser     string
	DatabasePassword     string
	DatabaseRootPassword string
	DatabaseTable        string
	DatabaseAddress      string
	Port                 string
}

// Get returns a config object that is built from environment variables
func Get() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	c := &Config{
		DatabaseUser:         os.Getenv("DB_USER"),
		DatabaseRootUser:     os.Getenv("DB_ROOT_USER"),
		DatabasePassword:     os.Getenv("DB_PASS"),
		DatabaseRootPassword: os.Getenv("DB_ROOT_PASS"),
		DatabaseTable:        os.Getenv("DB_TABLE"),
		DatabaseAddress:      os.Getenv("DB_ADDRESS"),
		Port:                 os.Getenv("PORT"),
	}

	return c, nil
}
