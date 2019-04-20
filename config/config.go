package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	DatabaseUser         string
	DatabaseRootUser     string
	DatabasePassword     string
	DatabaseRootPassword string
	DatabaseTable        string
	DatabaseAddress      string
	Port                 string
	DatabaseDebug        bool
}

// Get returns a config object that is built from environment variables
func Get() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	debug, err := strconv.ParseBool(getenv("DB_DEBUG", "0"))
	if err != nil {
		debug = false
	}

	c := &Config{
		DatabaseUser:         getenv("DB_USER", ""),
		DatabaseRootUser:     getenv("DB_ROOT_USER", ""),
		DatabasePassword:     getenv("DB_PASS", ""),
		DatabaseRootPassword: getenv("DB_ROOT_PASS", ""),
		DatabaseTable:        getenv("DB_TABLE", ""),
		DatabaseAddress:      getenv("DB_ADDRESS", ""),
		Port:                 getenv("PORT", ""),
		DatabaseDebug:        debug,
	}

	return c, nil
}

func getenv(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
