package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultPriority string   `yaml:"default_priority"`
	AutoStage       bool     `yaml:"auto_stage"`
	Labels          []string `yaml:"labels"`
}

func Default() *Config {
	return &Config{
		DefaultPriority: "medium",
		AutoStage:       true,
		Labels:          []string{"bug", "feature", "auth", "security", "docs"},
	}
}

func Load(issuesDir string) (*Config, error) {
	path := filepath.Join(issuesDir, ".config.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func WriteDefault(issuesDir string) error {
	cfg := Default()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(issuesDir, ".config.yml"), data, 0644)
}
