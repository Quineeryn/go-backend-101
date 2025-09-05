package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port  string
	Env   string
	DBDSN string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Logging
	LogFilePath   string
	LogMaxSizeMB  int
	LogMaxBackups int
	LogMaxAgeDays int
}

func FromEnv() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		Env:           getEnv("APP_ENV", "development"),
		DBDSN:         getEnv("DB_DSN", ""),
		LogFilePath:   getEnv("LOG_FILE", "./logs/app.log"),
		LogMaxSizeMB:  getEnvInt("LOG_MAX_SIZE", 10),
		LogMaxBackups: getEnvInt("LOG_MAX_BACKUPS", 5),
		LogMaxAgeDays: getEnvInt("LOG_MAX_AGE", 30),

		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
	}
}

func getEnv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		var out int
		_, err := fmt.Sscanf(v, "%d", &out)
		if err == nil {
			return out
		}
	}
	return def
}
