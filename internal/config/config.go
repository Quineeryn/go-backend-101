package config

import "os"

type Config struct {
	Port string
	Env  string
}

func FromEnv() Config {
	return Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),
	}
}

func getEnv(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}
