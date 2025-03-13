package config

import (
	"os"
	"strconv"
)

const (
	KeyDBConn         = "db/connection"
	KeyDBDatabase     = "db/database"
	KeyDBDMinPoolSize = "db/min_pool_size"
	KeyDBMaxPoolSize  = "db/max_pool_size"
	KeyDBDefaultSize  = "db/default_size"
	KeyDBMaxSize      = "db/max_size"

	DefaultDBConn        = "mongodb://localhost:27017/game2d"
	DefaultDBDatabase    = "game2d"
	DefaultDBMinPoolSize = 20
	DefaultDBMaxPoolSize = 100
	DefaultDBDefaultSize = 100
	DefaultDBMaxSize     = 10000
)

// DBConfig values represent database configuration data.
type DBConfig struct {
	Conn        string `json:"connection,omitempty"    yaml:"connection,omitempty"`
	Database    string `json:"database,omitempty"      yaml:"database,omitempty"`
	MinPoolSize int    `json:"min_pool_size,omitempty" yaml:"min_pool_size,omitempty"`
	MaxPoolSize int    `json:"max_pool_size,omitempty" yaml:"max_pool_size,omitempty"`
	DefaultSize int64  `json:"default_size,omitempty"  yaml:"default_size,omitempty"`
	MaxSize     int64  `json:"max_size,omitempty"      yaml:"max_size,omitempty"`
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

	if v := os.Getenv(ReplaceEnv(KeyDBDatabase)); v != "" {
		c.Database = v
	}

	if c.Database == "" {
		c.Database = DefaultDBDatabase
	}

	if v := os.Getenv(ReplaceEnv(KeyDBDMinPoolSize)); v != "" {
		v, err := strconv.Atoi(v)
		if err != nil {
			v = DefaultDBMinPoolSize
		}

		c.MinPoolSize = v
	}

	if c.MinPoolSize == 0 {
		c.MinPoolSize = DefaultDBMinPoolSize
	}

	if v := os.Getenv(ReplaceEnv(KeyDBMaxPoolSize)); v != "" {
		v, err := strconv.Atoi(v)
		if err != nil {
			v = DefaultDBMaxPoolSize
		}

		c.MaxPoolSize = v
	}

	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = DefaultDBMaxPoolSize
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

// DBDatabase returns the name of the database used by the primary database
// connection pool.
func (c *Config) DBDatabase() string {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBDatabase
	}

	return c.db.Database
}

// DBMinPoolSize returns the minimum number of connections in the database
// connection pool.
func (c *Config) DBMinPoolSize() int {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMinPoolSize
	}

	return c.db.MinPoolSize
}

// DBMaxPoolSize returns the maximum number of connections in the database
// connection pool.
func (c *Config) DBMaxPoolSize() int {
	c.RLock()
	defer c.RUnlock()

	if c.db == nil {
		return DefaultDBMaxPoolSize
	}

	return c.db.MaxPoolSize
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
