package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseUser     string
	DatabasePassword string
	DatabaseTable    string
	Port             string
}

// getConfig returns a config object that is built from environment variables
func getConfig() (*Config, error) {
	err := godotenv.Load()
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
