package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Env      string
}

type AppConfig struct {
	Name    string
	Version string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	ConnectionString string
	MaxOpenConns     int
	MaxIdleConns     int
}

type AuthConfig struct {
	APIKey string
}

var cfg *Config

func Init() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	envFile := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: Failed to load %s: %v", envFile, err)
		} else {
			log.Printf("Loaded config from: %s", envFile)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(",", "_"))

	cfg = &Config{
		Env: env,
		App: AppConfig{
			Name:    getEnv("APP_NAME", "kasir-api"),
			Version: getEnv("APP_VERSION", "1.0.0"),
		},

		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
		},

		Database: DatabaseConfig{
			ConnectionString: getEnv("DB_CONN", ""),
			MaxOpenConns:     getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:     getInt("DB_MAX_IDLE_CONNS", 5),
		},

		Auth: AuthConfig{
			APIKey: getEnv("API_KEY", ""),
		},
	}

	if cfg.Database.ConnectionString == "" {
		return nil, fmt.Errorf("DB_CONN is required")
	}

	if cfg.Auth.APIKey == "" && env == "production" {
		return nil, fmt.Errorf("API_KEY is required for production")
	}

	return cfg, nil

}

func Get() *Config {
	if cfg == nil {
		panic("config not initialized")
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}

	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}

	return defaultValue
}
