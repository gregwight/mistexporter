package config

import (
	"fmt"
	"os"
	"time"

	"github.com/gregwight/mistclient"
	"gopkg.in/yaml.v3"
)

const (
	defaultAPIURL                   string        = "https://api.mist.com"
	defaultExporterAddress          string        = "0.0.0.0"
	defaultExporterPort             int           = 10038
	defaultCollectTimeout           time.Duration = 30 * time.Second
	defaultSiteRefreshInterval      time.Duration = 1 * time.Minute
	defaultDeviceNameRefresInterval time.Duration = 1 * time.Minute
)

// Config holds the top-level exporter configuration.
type Config struct {
	OrgId      string             `yaml:"org_id,omitempty"`
	MistClient *mistclient.Config `yaml:"mist_api,omitempty"`
	Exporter   *Exporter          `yaml:"exporter,omitempty"`
	Collector  *Collector         `yaml:"collector,omitempty"`
}

// Exporter holds configurtion relevant to exporter's HTTP server.
type Exporter struct {
	Address string `yaml:"address,omitempty"`
	Port    int    `yaml:"port,omitempty"`
}

// Collector holds configuration relevant to metrics collection.
type Collector struct {
	CollectTimeout            time.Duration `yaml:"collect_timeout,omitempty"`
	DeviceNameRefreshInterval time.Duration `yaml:"device_name_refresh_interval,omitempty"`
	SiteRefreshInterval       time.Duration `yaml:"site_refresh_interval,omitempty"`
	SiteFilter                *SiteFilter   `yaml:"site_filter,omitempty"`
}

// SiteFilter defines rules for including or excluding sites from collection.
type SiteFilter struct {
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// LoadConfig loads and processes the YAML configuration with environment variable substitution.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Replace environment variables in the format ${VAR_NAME}
	configStr := string(data)
	configStr = os.ExpandEnv(configStr)

	config := newDefaultConfig()
	if err := yaml.Unmarshal([]byte(configStr), config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

func newDefaultConfig() *Config {
	return &Config{
		MistClient: &mistclient.Config{
			BaseURL: defaultAPIURL,
		},
		Exporter: &Exporter{
			Address: defaultExporterAddress,
			Port:    defaultExporterPort,
		},
		Collector: &Collector{
			CollectTimeout:            defaultCollectTimeout,
			DeviceNameRefreshInterval: defaultDeviceNameRefresInterval,
			SiteRefreshInterval:       defaultSiteRefreshInterval,
		},
	}
}
