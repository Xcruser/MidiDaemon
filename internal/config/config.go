// Package config verwaltet die Konfiguration für MidiDaemon.
// Es lädt und verarbeitet JSON-Mappings zwischen MIDI-Events und Systemaktionen.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config repräsentiert die Hauptkonfiguration von MidiDaemon
type Config struct {
	// MIDI-Einstellungen
	MIDI MIDIConfig `json:"midi"`

	// Mappings zwischen MIDI-Events und Aktionen
	Mappings []Mapping `json:"mappings"`

	// Allgemeine Einstellungen
	General GeneralConfig `json:"general"`
}

// MIDIConfig enthält MIDI-spezifische Einstellungen
type MIDIConfig struct {
	// Port-Name für MIDI-Eingabe (optional, verwendet ersten verfügbaren Port wenn leer)
	InputPort string `json:"input_port"`

	// MIDI-Kanal (0-15, -1 für alle Kanäle)
	Channel int `json:"channel"`

	// Timeout für MIDI-Verbindung in Sekunden
	Timeout int `json:"timeout"`
}

// Mapping definiert eine Zuordnung zwischen MIDI-Event und Systemaktion
type Mapping struct {
	// Name des Mappings (für Logging und Debugging)
	Name string `json:"name"`

	// MIDI-Event Definition
	Event MIDIEvent `json:"event"`

	// Systemaktion die ausgeführt werden soll
	Action Action `json:"action"`

	// Aktiviert/Deaktiviert
	Enabled bool `json:"enabled"`
}

// MIDIEvent definiert ein MIDI-Event
type MIDIEvent struct {
	// Typ des Events: "note_on", "note_off", "control_change", "program_change"
	Type string `json:"type"`

	// MIDI-Note (0-127) für Note Events
	Note int `json:"note,omitempty"`

	// Controller-Nummer (0-127) für Control Change Events
	Controller int `json:"controller,omitempty"`

	// Program-Nummer (0-127) für Program Change Events
	Program int `json:"program,omitempty"`

	// Velocity-Schwellwert für Note Events (0-127)
	Velocity int `json:"velocity,omitempty"`

	// Controller-Wert-Schwellwert für Control Change Events (0-127)
	Value int `json:"value,omitempty"`
}

// Action definiert eine Systemaktion
type Action struct {
	// Typ der Aktion: "volume", "app_start", "key_combination", "audio_source"
	Type string `json:"type"`

	// Parameter für die Aktion (abhängig vom Typ)
	Parameters map[string]interface{} `json:"parameters"`
}

// GeneralConfig enthält allgemeine Einstellungen
type GeneralConfig struct {
	// Log-Level: "debug", "info", "warn", "error"
	LogLevel string `json:"log_level"`

	// Automatischer Neustart bei Fehlern
	AutoRestart bool `json:"auto_restart"`

	// Verzögerung zwischen Aktionen in Millisekunden
	ActionDelay int `json:"action_delay"`
}

// Load lädt die Konfiguration aus einer JSON-Datei
func Load(configPath string) (*Config, error) {
	// Absoluten Pfad erstellen
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("ungültiger Konfigurationspfad: %w", err)
	}

	// Datei öffnen
	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("konnte Konfigurationsdatei nicht öffnen: %w", err)
	}
	defer file.Close()

	// JSON dekodieren
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("fehler beim Parsen der Konfigurationsdatei: %w", err)
	}

	// Standardwerte setzen falls nicht definiert
	setDefaults(&config)

	// Konfiguration validieren
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("ungültige Konfiguration: %w", err)
	}

	return &config, nil
}

// setDefaults setzt Standardwerte für fehlende Konfigurationsoptionen
func setDefaults(config *Config) {
	// MIDI-Standardwerte
	if config.MIDI.Channel == 0 {
		config.MIDI.Channel = -1 // Alle Kanäle
	}
	if config.MIDI.Timeout == 0 {
		config.MIDI.Timeout = 30 // 30 Sekunden
	}

	// Allgemeine Standardwerte
	if config.General.LogLevel == "" {
		config.General.LogLevel = "info"
	}
	if config.General.ActionDelay == 0 {
		config.General.ActionDelay = 100 // 100ms
	}

	// Alle Mappings standardmäßig aktivieren
	for i := range config.Mappings {
		if !config.Mappings[i].Enabled {
			config.Mappings[i].Enabled = true
		}
	}
}

// validate überprüft die Konfiguration auf Gültigkeit
func validate(config *Config) error {
	// MIDI-Kanal validieren
	if config.MIDI.Channel < -1 || config.MIDI.Channel > 15 {
		return fmt.Errorf("ungültiger MIDI-Kanal: %d (muss zwischen -1 und 15 liegen)", config.MIDI.Channel)
	}

	// Mappings validieren
	for i, mapping := range config.Mappings {
		if err := validateMapping(&mapping); err != nil {
			return fmt.Errorf("ungültiges Mapping %d (%s): %w", i, mapping.Name, err)
		}
	}

	return nil
}

// validateMapping überprüft ein einzelnes Mapping auf Gültigkeit
func validateMapping(mapping *Mapping) error {
	// Event validieren
	if err := validateMIDIEvent(&mapping.Event); err != nil {
		return fmt.Errorf("ungültiges MIDI-Event: %w", err)
	}

	// Action validieren
	if err := validateAction(&mapping.Action); err != nil {
		return fmt.Errorf("ungültige Aktion: %w", err)
	}

	return nil
}

// validateMIDIEvent überprüft ein MIDI-Event auf Gültigkeit
func validateMIDIEvent(event *MIDIEvent) error {
	switch event.Type {
	case "note_on", "note_off":
		if event.Note < 0 || event.Note > 127 {
			return fmt.Errorf("ungültige MIDI-Note: %d (muss zwischen 0 und 127 liegen)", event.Note)
		}
		if event.Velocity < 0 || event.Velocity > 127 {
			return fmt.Errorf("ungültige Velocity: %d (muss zwischen 0 und 127 liegen)", event.Velocity)
		}
	case "control_change":
		if event.Controller < 0 || event.Controller > 127 {
			return fmt.Errorf("ungültiger Controller: %d (muss zwischen 0 und 127 liegen)", event.Controller)
		}
		if event.Value < 0 || event.Value > 127 {
			return fmt.Errorf("ungültiger Controller-Wert: %d (muss zwischen 0 und 127 liegen)", event.Value)
		}
	case "program_change":
		if event.Program < 0 || event.Program > 127 {
			return fmt.Errorf("ungültiges Program: %d (muss zwischen 0 und 127 liegen)", event.Program)
		}
	default:
		return fmt.Errorf("ungültiger Event-Typ: %s", event.Type)
	}

	return nil
}

// validateAction überprüft eine Aktion auf Gültigkeit
func validateAction(action *Action) error {
	switch action.Type {
	case "volume":
		// Volume-Aktionen benötigen einen "direction" Parameter
		if _, ok := action.Parameters["direction"]; !ok {
			return fmt.Errorf("volume-Aktion benötigt 'direction' Parameter")
		}
	case "app_start":
		// App-Start-Aktionen benötigen einen "path" Parameter
		if _, ok := action.Parameters["path"]; !ok {
			return fmt.Errorf("app_start-Aktion benötigt 'path' Parameter")
		}
	case "key_combination":
		// Tastenkombinationen benötigen "keys" Parameter
		if _, ok := action.Parameters["keys"]; !ok {
			return fmt.Errorf("key_combination-Aktion benötigt 'keys' Parameter")
		}
	case "audio_source":
		// Audio-Quelle benötigt "source" Parameter
		if _, ok := action.Parameters["source"]; !ok {
			return fmt.Errorf("audio_source-Aktion benötigt 'source' Parameter")
		}
	default:
		return fmt.Errorf("ungültiger Aktion-Typ: %s", action.Type)
	}

	return nil
}

// Save speichert die Konfiguration in eine JSON-Datei
func Save(configPath string, config *Config) error {
	// Absoluten Pfad erstellen
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("ungültiger Konfigurationspfad: %w", err)
	}

	// Datei erstellen/überschreiben
	file, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("konnte Konfigurationsdatei nicht erstellen: %w", err)
	}
	defer file.Close()

	// JSON kodieren
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("fehler beim Schreiben der Konfigurationsdatei: %w", err)
	}

	return nil
}
