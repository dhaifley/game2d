package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/game2d/config"
)

func TestServiceConfig(t *testing.T) {
	t.Parallel()

	cfg := config.New("test name")

	cfg.Load(nil)

	cfg.SetService(&config.ServiceConfig{
		Name:             "test name",
		Maintenance:      true,
		ImportInterval:   time.Second,
		GameLimitDefault: 5,
	})

	if cfg.ServiceName() != "test name" {
		t.Errorf("Expected name: test name, got: %v", cfg.ServiceName())
	}

	if cfg.ServiceMaintenance() != true {
		t.Errorf("Expected maintenance: true, got: %v",
			cfg.ServiceMaintenance())
	}

	if cfg.ImportInterval() != time.Second {
		t.Errorf("Expected import interval: 1s, got: %v", cfg.ImportInterval())
	}

	if cfg.GameLimitDefault() != 5 {
		t.Errorf("Expected game limit default: 5, got: %v",
			cfg.GameLimitDefault())
	}
}
