package config

import (
	"os"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	StoragePath  string
	PublicDomain  string
	AdminPassword string
	Port          string
}

func Load() *Config {
	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "charon"),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "charon"),
		StoragePath:  getEnv("STORAGE_PATH", "/data/uploads"),
		PublicDomain:  getEnv("PUBLIC_DOMAIN", "localhost"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		Port:          getEnv("PORT", "8080"),
	}
}

func (c *Config) DSN() string {
	return c.DBUser + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBName + "?parseTime=true&charset=utf8mb4"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
