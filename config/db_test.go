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
		DefaultSize: 10,
		MaxSize:     100,
	})

	if cfg.DBConn() != exp {
		t.Errorf("Expected connection: %v, got: %v", exp, cfg.DBConn())
	}

	if cfg.DBDefaultSize() != 10 {
		t.Errorf("Expected default size: 10, got: %v", cfg.DBDefaultSize())
	}

	if cfg.DBMaxSize() != 100 {
		t.Errorf("Expected max size: 100, got: %v", cfg.DBMaxSize())
	}
}
