package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Setup: Create a temporary config file
	content := `
org_id: "my-org-id"
mist_api:
  base_url: "https://api.mist.com"
  api_key: "my-api-key"
  timeout: 15s
exporter:
  address: "127.0.0.1"
  port: 9999
collector:
  collect_timeout: 25s
  site_refresh_interval: 5m
  site_filter:
    include: ["Main Office-*"]
    exclude: ["Main Office-Guest"]
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	// Test loading the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned an unexpected error: %v", err)
	}

	if cfg.OrgId != "my-org-id" {
		t.Errorf("expected OrgId to be 'my-org-id', got %q", cfg.OrgId)
	}
	if cfg.MistClient.BaseURL != "https://api.mist.com" {
		t.Errorf("expected BaseURL to be 'https://api.mist.com', got %q", cfg.MistClient.BaseURL)
	}
	if cfg.MistClient.APIKey != "my-api-key" {
		t.Errorf("expected APIKey to be 'my-api-key', got %q", cfg.MistClient.APIKey)
	}
	if cfg.MistClient.Timeout != (15 * time.Second) {
		t.Errorf("expected Timeout to be 15s, got %v", cfg.MistClient.Timeout)
	}
	if cfg.Exporter.Address != "127.0.0.1" {
		t.Errorf("expected Address to be '127.0.0.1', got %q", cfg.Exporter.Address)
	}
	if cfg.Exporter.Port != 9999 {
		t.Errorf("expected Port to be 9999, got %d", cfg.Exporter.Port)
	}
	if cfg.Collector.CollectTimeout != 25*time.Second {
		t.Errorf("expected Collector.Timeout to be 25s, got %v", cfg.Collector.CollectTimeout)
	}
	if cfg.Collector.SiteRefreshInterval != 5*time.Minute {
		t.Errorf("expected Collector.SiteRefreshInterval to be 5m, got %v", cfg.Collector.SiteRefreshInterval)
	}
	if cfg.Collector.SiteFilter == nil {
		t.Fatal("expected SiteFilter to be loaded, but it was nil")
	}
	if len(cfg.Collector.SiteFilter.Include) != 1 || cfg.Collector.SiteFilter.Include[0] != "Main Office-*" {
		t.Errorf("unexpected SiteFilter.Include: got %v", cfg.Collector.SiteFilter.Include)
	}
	if len(cfg.Collector.SiteFilter.Exclude) != 1 || cfg.Collector.SiteFilter.Exclude[0] != "Main Office-Guest" {
		t.Errorf("unexpected SiteFilter.Exclude: got %v", cfg.Collector.SiteFilter.Exclude)
	}
}

func TestLoadConfig_WithEnvVars(t *testing.T) {
	// Setup: Set environment variables
	t.Setenv("TEST_MIST_ORG_ID", "env-org-id")
	t.Setenv("TEST_MIST_API_KEY", "env-api-key")

	content := `
org_id: "${TEST_MIST_ORG_ID}"
mist_api:
  api_key: "${TEST_MIST_API_KEY}"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	// Test loading the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned an unexpected error: %v", err)
	}

	if cfg.OrgId != "env-org-id" {
		t.Errorf("expected OrgId to be 'env-org-id', got %q", cfg.OrgId)
	}
	if cfg.MistClient.APIKey != "env-api-key" {
		t.Errorf("expected APIKey to be 'env-api-key', got %q", cfg.MistClient.APIKey)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Setup: Create a minimal config file
	content := `
mist_api:
  api_key: "my-api-key"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	// Test loading the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() returned an unexpected error: %v", err)
	}

	// Check that defaults are applied
	if cfg.Exporter.Address != defaultExporterAddress {
		t.Errorf("expected default Address to be %q, got %q", defaultExporterAddress, cfg.Exporter.Address)
	}
	if cfg.Exporter.Port != defaultExporterPort {
		t.Errorf("expected default Port to be %d, got %d", defaultExporterPort, cfg.Exporter.Port)
	}
	if cfg.Collector.CollectTimeout != defaultCollectTimeout {
		t.Errorf("expected default Collector.Timeout to be %v, got %v", defaultCollectTimeout, cfg.Collector.CollectTimeout)
	}
	if cfg.Collector.DeviceNameRefreshInterval != defaultDeviceNameRefresInterval {
		t.Errorf("expected default Collector.DeviceNameRefreshInterval to be %v, got %v", defaultDeviceNameRefresInterval, cfg.Collector.DeviceNameRefreshInterval)
	}
	if cfg.Collector.SiteRefreshInterval != defaultSiteRefreshInterval {
		t.Errorf("expected default Collector.SiteRefreshInterval to be %v, got %v", defaultSiteRefreshInterval, cfg.Collector.SiteRefreshInterval)
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("expected an error for non-existent file, but got nil")
	}
}
