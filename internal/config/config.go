package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBUrl     string
	RedisUrl  string
	JwtSecret string
	Port      string
}

func LoadConfig() (*Config, error) {
	// Set the config file name and type
	viper.SetConfigFile(".env")
	viper.SetConfigType("dotenv")

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		// If the .env file doesn't exist, continue with environment variables only
		// This allows the app to work in production without a .env file
	}

	// Also read from environment variables (this takes precedence over .env file)
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("DB_URL", "postgres://user:password@localhost:5432/evently_db?sslmode=disable")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production")
	viper.SetDefault("PORT", "8080")

	cfg := &Config{
		DBUrl:     viper.GetString("DB_URL"),
		RedisUrl:  viper.GetString("REDIS_URL"),
		JwtSecret: viper.GetString("JWT_SECRET"),
		Port:      viper.GetString("PORT"),
	}

	// Validate required config
	if cfg.JwtSecret == "" {
		cfg.JwtSecret = "fallback-secret-key"
	}

	return cfg, nil
}

// GetPort returns the port with colon prefix for server binding
func (c *Config) GetPort() string {
	if c.Port == "" {
		return ":8080"
	}
	if c.Port[0] != ':' {
		return ":" + c.Port
	}
	return c.Port
}
