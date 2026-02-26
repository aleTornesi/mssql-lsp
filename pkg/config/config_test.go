package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}
	if cfg.LowercaseKeywords {
		t.Error("LowercaseKeywords should default to false")
	}
	if len(cfg.Connections) != 0 {
		t.Error("Connections should default to empty")
	}
}

func TestValidate(t *testing.T) {
	cfg := NewConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}
