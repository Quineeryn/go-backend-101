package config

import "os"

type Config struct {
	Port  string
	Env   string
	DBDSN string
}

func FromEnv() Config {
	return Config{
		Port:  getEnv("PORT", "8080"),
		Env:   getEnv("APP_ENV", "development"),
		DBDSN: getEnv("DB_DSN", ""),
	}
}

func getEnv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}
