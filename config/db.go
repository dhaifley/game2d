package config

import (
	"os"
	"strconv"
)

const (
	KeyDBConn        = "db/connection"
	KeyDBDefaultSize = "db/default_size"
	KeyDBMaxSize     = "db/max_size"

	DefaultDBConn        = "mongodb://localhost:27017"
	DefaultDBDefaultSize = 100
	DefaultDBMaxSize     = 10000
)

const (
	DBModeNormal = iota
	DBModeMigrate
	DBModeInit
)

// DBConfig values represent database configuration data.
type DBConfig struct {
	Conn        string `json:"connection,omitempty"   yaml:"connection,omitempty"`
	DefaultSize int64  `json:"default_size,omitempty" yaml:"default_size,omitempty"`
	MaxSize     int64  `json:"max_size,omitempty"     yaml:"max_size,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *DBConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyDBConn)); v != "" {
		c.Conn = v
	}

	if c.Conn == "" {
		c.Conn = DefaultDBConn
	}

	if v := os.Getenv(ReplaceEnv(KeyDBDefaultSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultDBDefaultSize
		}

		c.DefaultSize = v
	}

	if c.DefaultSize == 0 {
		c.DefaultSize = DefaultDBDefaultSize
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMaxSize)); v != "" {
		v, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			v = DefaultDBMaxSize
		}

		c.MaxSize = v
	}

	if c.MaxSize == 0 {
		c.MaxSize = DefaultDBMaxSize
	}
}

// DBConn returns the connection string used by the primary database
// connection pool.
func (c *Config) DBConn() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBConn
	}

	return c.db.Conn
}

// DBDefaultSize returns the default number of rows any query will return.
func (c *Config) DBDefaultSize() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBDefaultSize
	}

	return c.db.DefaultSize
}

// DBMaxSize returns the maximum limit of rows any query may return.
func (c *Config) DBMaxSize() int64 {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMaxSize
	}

	return c.db.MaxSize
}
