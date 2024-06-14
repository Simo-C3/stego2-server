package config

import (
	"cmp"
	"os"
)

type Config struct {
	ServerPort string
}

func New() *Config {
	return &Config{
		ServerPort: loadEnv("PORT", "8080"),
	}
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:     loadEnv("DB_HOST", "localhost"),
		Port:     loadEnv("DB_PORT", "3306"),
		User:     loadEnv("DB_USER", "root"),
		Password: loadEnv("DB_PASSWORD", "password"),
		DBName:   loadEnv("DB_NAME", "main"),
	}
}

func loadEnv(env, def string) string {
	return cmp.Or(os.Getenv(env), def)
}
