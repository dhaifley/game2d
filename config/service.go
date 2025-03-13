package config

import (
	"os"
	"strconv"
	"time"
)

const (
	KeyServiceName        = "service/name"
	KeyServiceMaintenance = "service/maintenance"
	KeyImportInterval     = "service/import_interval"
	KeyGameDataRetention  = "game/data_retention"

	DefaultServiceName        = "game2d-api"
	DefaultServiceMaintenance = false
	DefaultImportInterval     = time.Minute * 5
	DefaultGameDataRetention  = time.Hour * 720 // 30d
)

// ServiceConfig values represent telemetry configuration data.
type ServiceConfig struct {
	Name              string        `json:"name,omitempty"                yaml:"name,omitempty"`
	Maintenance       bool          `json:"maintenance,omitempty"         yaml:"maintenance,omitempty"`
	ImportInterval    time.Duration `json:"import_interval,omitempty"     yaml:"import_interval,omitempty"`
	GameDataRetention time.Duration `json:"game_data_retention,omitempty" yaml:"game_data_retention,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *ServiceConfig) Load() {
	if c.Name == "" {
		c.Name = DefaultServiceName
	}

	if v := os.Getenv(ReplaceEnv(KeyServiceMaintenance)); v != "" {
		v, err := strconv.ParseBool(v)
		if err != nil {
			v = DefaultServiceMaintenance
		}

		c.Maintenance = v
	}

	if v := os.Getenv(ReplaceEnv(KeyImportInterval)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultImportInterval
		}

		c.ImportInterval = v
	}

	if c.ImportInterval == 0 {
		c.ImportInterval = DefaultImportInterval
	}

	if v := os.Getenv(ReplaceEnv(KeyGameDataRetention)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultGameDataRetention
		}

		c.GameDataRetention = v
	}

	if c.GameDataRetention == 0 {
		c.GameDataRetention = DefaultGameDataRetention
	}
}

// ServiceName returns the name of the service.
func (c *Config) ServiceName() string {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultServiceName
	}

	return c.service.Name
}

// ServiceMaintenance returns whether the service has been placed into
// maintenance mode.
func (c *Config) ServiceMaintenance() bool {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultServiceMaintenance
	}

	return c.service.Maintenance
}

// ImportInterval returns the frequency at which repository imports are
// performed.
func (c *Config) ImportInterval() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultImportInterval
	}

	return c.service.ImportInterval
}

// GameDataRetention returns the duration for which game data elements are
// retained. Default value is 30 days.
func (c *Config) GameDataRetention() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultGameDataRetention
	}

	return c.service.GameDataRetention
}
