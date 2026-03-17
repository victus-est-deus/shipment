package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type StorageType string

const (
	StoragePostgres StorageType = "postgres"
	StorageJSON     StorageType = "json"
)

const DefaultStoragePath = "storage/tmp"

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Storage  StorageType    `yaml:"-"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultConfig()
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg.Storage = StoragePostgres
	return &cfg, nil
}

func defaultConfig() (*Config, error) {
	dirs := []string{
		DefaultStoragePath + "/shipments",
		DefaultStoragePath + "/status_events",
		DefaultStoragePath + "/logs",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating storage directory %s: %w", dir, err)
		}
	}

	return &Config{
		GRPC:    GRPCConfig{Port: 50051},
		Storage: StorageJSON,
	}, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}
