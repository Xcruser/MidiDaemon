package config

import (
	"encoding/json"
	"testing"
)

func TestSetDefaultsChannelUnset(t *testing.T) {
	data := []byte(`{"midi": {"timeout": 5}, "mappings": [], "general": {}}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	setDefaults(&cfg)
	if cfg.MIDI.Channel != -1 {
		t.Fatalf("expected channel -1, got %d", cfg.MIDI.Channel)
	}
}

func TestSetDefaultsChannelZero(t *testing.T) {
	data := []byte(`{"midi": {"channel": 0}, "mappings": [], "general": {}}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	setDefaults(&cfg)
	if cfg.MIDI.Channel != 0 {
		t.Fatalf("expected channel 0, got %d", cfg.MIDI.Channel)
	}
}
