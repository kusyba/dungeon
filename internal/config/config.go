package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Floors < 1 {
		return fmt.Errorf("Floors must be >= 1, got %d", c.Floors)
	}
	if c.Monsters < 0 {
		return fmt.Errorf("Monsters must be >= 0, got %d", c.Monsters)
	}
	if c.Duration < 0 {
		return fmt.Errorf("Duration must be >= 0, got %d", c.Duration)
	}
	if _, err := time.Parse("15:04:05", c.OpenAt); err != nil {
		return fmt.Errorf("invalid OpenAt format (expected HH:MM:SS): %w", err)
	}
	return nil
}

func (c *Config) GetOpenTime() (time.Time, error) {
	return time.Parse("15:04:05", c.OpenAt)
}

func (c *Config) GetCloseTime(openTime time.Time) time.Time {
	return openTime.Add(time.Duration(c.Duration) * time.Hour)
}
