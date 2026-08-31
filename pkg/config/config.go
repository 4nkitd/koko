package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type EngineConfig struct {
	Model string  `json:"model"`
	Voice string  `json:"voice"`
	Speed float64 `json:"speed"`
}

type Config struct {
	DefaultMode string       `json:"default_mode"`
	Monotone    EngineConfig `json:"monotone"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultMode: "native",
		Monotone: EngineConfig{
			Model: "kokoro-en-v0_19",
			Voice: "ironman",
			Speed: 1.0,
		},
	}
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".koko-config.json"
	}
	return filepath.Join(home, ".config", "koko", "config.json")
}

func LoadConfig(customPath string) (*Config, error) {
	cfgPath := customPath
	if cfgPath == "" {
		cfgPath = GetConfigPath()
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			_ = SaveConfig(cfg, cfgPath)
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return cfg, nil
}

func SaveConfig(cfg *Config, customPath string) error {
	cfgPath := customPath
	if cfgPath == "" {
		cfgPath = GetConfigPath()
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	return os.WriteFile(cfgPath, data, 0644)
}
