package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.DefaultPriority != "medium" {
		t.Errorf("expected default priority 'medium', got %q", cfg.DefaultPriority)
	}
	if !cfg.AutoStage {
		t.Error("expected auto_stage to be true by default")
	}
	if len(cfg.Labels) == 0 {
		t.Error("expected default labels to be non-empty")
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultPriority != "medium" {
		t.Errorf("expected fallback to default priority, got %q", cfg.DefaultPriority)
	}
}

func TestLoadExistingFile(t *testing.T) {
	dir := t.TempDir()
	content := `default_priority: high
auto_stage: false
labels:
  - bug
  - feature
`
	if err := os.WriteFile(filepath.Join(dir, ".config.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultPriority != "high" {
		t.Errorf("expected priority 'high', got %q", cfg.DefaultPriority)
	}
	if cfg.AutoStage {
		t.Error("expected auto_stage to be false")
	}
	if len(cfg.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(cfg.Labels))
	}
}

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	if err := WriteDefault(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultPriority != "medium" {
		t.Errorf("expected priority 'medium', got %q", cfg.DefaultPriority)
	}
	if !cfg.AutoStage {
		t.Error("expected auto_stage to be true")
	}
}
