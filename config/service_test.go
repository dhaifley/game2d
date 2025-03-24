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
		Name:              "test name",
		AccountID:         "test id",
		AccountName:       "test name",
		Maintenance:       true,
		ImportInterval:    time.Second,
		GameLimitDefault:  5,
		PromptHistorySize: 10,
	})

	if cfg.ServiceName() != "test name" {
		t.Errorf("Expected name: test name, got: %v", cfg.ServiceName())
	}

	if cfg.AccountID() != "test id" {
		t.Errorf("Expected account id: test id, got: %v",
			cfg.AccountID())
	}

	if cfg.AccountName() != "test name" {
		t.Errorf("Expected account name: test name, got: %v",
			cfg.AccountName())
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

	if cfg.PromptHistorySize() != 10 {
		t.Errorf("Expected prompt history size: 10, got: %v",
			cfg.PromptHistorySize())
	}
}
