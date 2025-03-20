package config

import (
	"os"
	"strconv"
	"time"
)

const (
	KeyServiceName        = "service/name"
	KeyAccountID          = "account_id"
	KeyAccountName        = "account_name"
	KeyServiceMaintenance = "service/maintenance"
	KeyImportInterval     = "service/import_interval"
	KeyGameLimitDefault   = "service/game_limit_default"

	DefaultServiceName        = "game2d-api"
	DefaultAccountID          = "game2d"
	DefaultAccountName        = "game2d-api"
	DefaultServiceMaintenance = false
	DefaultImportInterval     = time.Minute * 5
	DefaultGameLimitDefault   = 10
)

// ServiceConfig values represent telemetry configuration data.
type ServiceConfig struct {
	Name             string        `json:"name,omitempty"               yaml:"name,omitempty"`
	AccountID        string        `json:"account_id,omitempty"         yaml:"account_id,omitempty"`
	AccountName      string        `json:"account_name,omitempty"       yaml:"account_name,omitempty"`
	Maintenance      bool          `json:"maintenance,omitempty"        yaml:"maintenance,omitempty"`
	ImportInterval   time.Duration `json:"import_interval,omitempty"    yaml:"import_interval,omitempty"`
	GameLimitDefault int64         `json:"game_limit_default,omitempty" yaml:"game_limit_default,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *ServiceConfig) Load() {
	if c.Name == "" {
		c.Name = DefaultServiceName
	}

	if v := os.Getenv(ReplaceEnv(KeyAccountID)); v != "" {
		c.AccountID = v
	}

	if c.AccountID == "" {
		c.AccountID = DefaultAccountID
	}

	if v := os.Getenv(ReplaceEnv(KeyAccountName)); v != "" {
		c.AccountName = v
	}

	if c.AccountName == "" {
		c.AccountName = DefaultAccountName
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

	if v := os.Getenv(ReplaceEnv(KeyGameLimitDefault)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultGameLimitDefault
		}

		c.GameLimitDefault = v
	}

	if c.GameLimitDefault == 0 {
		c.GameLimitDefault = DefaultGameLimitDefault
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

// AccountID returns the ID of the service account.
func (c *Config) AccountID() string {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultAccountID
	}

	return c.service.AccountID
}

// AccountName returns the name of the service account.
func (c *Config) AccountName() string {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultAccountName
	}

	return c.service.AccountName
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

// GameLimitDefault returns the default game limit for accounts.
func (c *Config) GameLimitDefault() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.service == nil {
		return DefaultGameLimitDefault
	}

	return c.service.GameLimitDefault
}
