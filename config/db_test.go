package config_test

import (
	"testing"

	"github.com/dhaifley/game2d/config"
)

func TestDatabaseConfig(t *testing.T) {
	t.Parallel()

	exp := "test"

	cfg := config.New("")

	cfg.Load(nil)

	cfg.SetDB(&config.DBConfig{
		Conn:        exp,
		Database:    exp,
		MinPoolSize: 1,
		MaxPoolSize: 10,
		DefaultSize: 10,
		MaxSize:     100,
	})

	if cfg.DBConn() != exp {
		t.Errorf("Expected connection: %v, got: %v", exp, cfg.DBConn())
	}

	if cfg.DBDatabase() != exp {
		t.Errorf("Expected database: %v, got: %v", exp, cfg.DBDatabase())
	}

	if cfg.DBMinPoolSize() != 1 {
		t.Errorf("Expected min pool size: 1, got: %v", cfg.DBMinPoolSize())
	}

	if cfg.DBMaxPoolSize() != 10 {
		t.Errorf("Expected max pool size: 10, got: %v", cfg.DBMaxPoolSize())
	}

	if cfg.DBDefaultSize() != 10 {
		t.Errorf("Expected default size: 10, got: %v", cfg.DBDefaultSize())
	}

	if cfg.DBMaxSize() != 100 {
		t.Errorf("Expected max size: 100, got: %v", cfg.DBMaxSize())
	}
}
