package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	PoolSize int
}

func Load() (*Config, error) {
	poolSize, _ := strconv.Atoi(getEnv("SMTP_POOL_SIZE", "5"))
	
	config := &Config{
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_FROM", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			PoolSize: poolSize,
		},
	}
	
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	return config, nil
}

func (c *Config) Validate() error {
	if c.SMTP.Host == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if c.SMTP.Username == "" {
		return fmt.Errorf("SMTP_FROM is required")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
