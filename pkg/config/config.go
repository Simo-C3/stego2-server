package config

import "os"

type Config struct {
	ServerPort string
	IsLocal    bool
}

func New() *Config {
	return &Config{
		ServerPort: os.Getenv("PORT"),
		IsLocal:    os.Getenv("IS_LOCAL") == "true",
	}
}


type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	IsLocal  bool
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Name:     os.Getenv("POSTGRES_DB"),
		IsLocal:  os.Getenv("IS_LOCAL") == "true",
	}
}
