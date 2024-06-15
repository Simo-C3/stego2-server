package config

import (
	"cmp"
	"os"
)

type Config struct {
	ServerPort string
	ENV        string
}

func New() *Config {
	return &Config{
		ServerPort: loadEnv("PORT", "8080"),
	}
}

type DBConfig struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	DBName                 string
	InstanceConnectionName string
	Env                    string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:                   loadEnv("DB_HOST", "localhost"),
		Port:                   loadEnv("DB_PORT", "3306"),
		User:                   loadEnv("DB_USER", "root"),
		Password:               loadEnv("DB_PASSWORD", "password"),
		DBName:                 loadEnv("DB_NAME", "main"),
		InstanceConnectionName: loadEnv("INSTANCE_CONNECTION_NAME", "default"),
		Env:                    loadEnv("ENV", "development"),
	}
}

type RedisConfig struct {
	Host string
	Port string
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host: loadEnv("REDIS_HOST", "localhost"),
		Port: loadEnv("REDIS_PORT", "6379"),
	}
}

func loadEnv(env, def string) string {
	return cmp.Or(os.Getenv(env), def)
}

type FirebaseConfig struct {
	ServiceAccount string
}

func NewFirebaseConfig() *FirebaseConfig {
	return &FirebaseConfig{
		ServiceAccount: os.Getenv("FIREBASE_SERVICE_ACCOUNT"),
	}
}
