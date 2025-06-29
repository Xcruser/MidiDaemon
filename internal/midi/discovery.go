// Package midi verwaltet MIDI-Eingaben und leitet sie an die entsprechenden Aktionen weiter.
// Diese Datei enthält die automatische MIDI-Controller-Erkennung.

package midi

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ControllerInfo enthält Informationen über einen erkannten MIDI-Controller
type ControllerInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	Type         ControllerType         `json:"type"`
	Capabilities ControllerCapabilities `json:"capabilities"`
	PortName     string                 `json:"port_name"`
	IsConnected  bool                   `json:"is_connected"`
	LastSeen     time.Time              `json:"last_seen"`
	Settings     ControllerSettings     `json:"settings"`
}

// ControllerType definiert den Typ des Controllers
type ControllerType string

const (
	ControllerTypeKeyboard    ControllerType = "keyboard"
	ControllerTypePad         ControllerType = "pad"
	ControllerTypeKnob        ControllerType = "knob"
	ControllerTypeSlider      ControllerType = "slider"
	ControllerTypeMixer       ControllerType = "mixer"
	ControllerTypeDJ          ControllerType = "dj"
	ControllerTypeDrumMachine ControllerType = "drum_machine"
	ControllerTypeSynthesizer ControllerType = "synthesizer"
	ControllerTypeUnknown     ControllerType = "unknown"
)

// ControllerCapabilities beschreibt die Fähigkeiten eines Controllers
type ControllerCapabilities struct {
	NumKeys       int  `json:"num_keys"`       // Anzahl der Tasten
	NumKnobs      int  `json:"num_knobs"`      // Anzahl der Drehregler
	NumSliders    int  `json:"num_sliders"`    // Anzahl der Schieberegler
	NumPads       int  `json:"num_pads"`       // Anzahl der Pads
	NumFaders     int  `json:"num_faders"`     // Anzahl der Fader
	HasDisplay    bool `json:"has_display"`    // Hat Display
	HasTransport  bool `json:"has_transport"`  // Hat Transport-Controls
	HasModWheel   bool `json:"has_mod_wheel"`  // Hat Modulationsrad
	HasPitchBend  bool `json:"has_pitch_bend"` // Hat Pitch Bend
	HasAftertouch bool `json:"has_aftertouch"` // Hat Aftertouch
	HasSustain    bool `json:"has_sustain"`    // Hat Sustain-Pedal
	HasExpression bool `json:"has_expression"` // Hat Expression-Pedal
}

// ControllerSettings enthält gerätespezifische Einstellungen
type ControllerSettings struct {
	DefaultChannel    int               `json:"default_channel"`    // Standard-MIDI-Kanal
	VelocitySensitive bool              `json:"velocity_sensitive"` // Velocity-sensitiv
	PressureSensitive bool              `json:"pressure_sensitive"` // Druck-sensitiv
	AutoMap           bool              `json:"auto_map"`           // Automatisches Mapping
	CustomMappings    map[string]string `json:"custom_mappings"`    // Benutzerdefinierte Mappings
}

// DiscoveryManager verwaltet die automatische MIDI-Controller-Erkennung
type DiscoveryManager struct {
	knownControllers map[string]*ControllerInfo
	portManager      MIDIPort
	logger           Logger
}

// Logger definiert die Schnittstelle für Logging
type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewDiscoveryManager erstellt einen neuen Discovery-Manager
func NewDiscoveryManager(portManager MIDIPort, logger Logger) *DiscoveryManager {
	return &DiscoveryManager{
		knownControllers: make(map[string]*ControllerInfo),
		portManager:      portManager,
		logger:           logger,
	}
}

// DiscoverControllers erkennt alle angeschlossenen MIDI-Controller
func (dm *DiscoveryManager) DiscoverControllers() ([]*ControllerInfo, error) {
	dm.logger.Info("Starte MIDI-Controller-Erkennung...")

	// Verfügbare MIDI-Ports abrufen
	portNames, err := dm.portManager.GetPortNames()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der MIDI-Ports: %w", err)
	}

	var controllers []*ControllerInfo

	// Jeden Port analysieren
	for _, portName := range portNames {
		controller, err := dm.analyzePort(portName)
		if err != nil {
			dm.logger.Warn("Fehler beim Analysieren des Ports", "port", portName, "error", err)
			continue
		}

		if controller != nil {
			controllers = append(controllers, controller)
			dm.knownControllers[controller.ID] = controller
		}
	}

	dm.logger.Info("MIDI-Controller-Erkennung abgeschlossen", "found", len(controllers))
	return controllers, nil
}

// analyzePort analysiert einen einzelnen MIDI-Port
func (dm *DiscoveryManager) analyzePort(portName string) (*ControllerInfo, error) {
	dm.logger.Debug("Analysiere MIDI-Port", "port", portName)

	// Controller-Info aus Port-Namen extrahieren
	controller := dm.extractControllerInfo(portName)
	if controller == nil {
		return nil, nil // Kein bekannter Controller
	}

	// Port öffnen und testen
	if err := dm.testPort(portName, controller); err != nil {
		dm.logger.Warn("Port-Test fehlgeschlagen", "port", portName, "error", err)
		return nil, err
	}

	controller.PortName = portName
	controller.IsConnected = true
	controller.LastSeen = time.Now()

	dm.logger.Info("Controller erkannt",
		"name", controller.Name,
		"manufacturer", controller.Manufacturer,
		"model", controller.Model,
		"type", controller.Type,
	)

	return controller, nil
}

// extractControllerInfo extrahiert Controller-Informationen aus dem Port-Namen
func (dm *DiscoveryManager) extractControllerInfo(portName string) *ControllerInfo {
	// Bekannte Controller-Patterns
	patterns := map[string]*ControllerInfo{
		// Akai Professional
		`(?i)akai.*mpk`: {
			Manufacturer: "Akai Professional",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      49, // MPK Mini
				NumKnobs:     8,
				NumPads:      8,
				HasTransport: true,
			},
		},
		`(?i)akai.*mpx`: {
			Manufacturer: "Akai Professional",
			Type:         ControllerTypePad,
			Capabilities: ControllerCapabilities{
				NumPads:    16,
				HasDisplay: true,
			},
		},

		// Native Instruments
		`(?i)traktor.*kontrol`: {
			Manufacturer: "Native Instruments",
			Type:         ControllerTypeDJ,
			Capabilities: ControllerCapabilities{
				NumKnobs:     8,
				NumSliders:   4,
				NumPads:      8,
				HasDisplay:   true,
				HasTransport: true,
			},
		},
		`(?i)maschine`: {
			Manufacturer: "Native Instruments",
			Type:         ControllerTypeDrumMachine,
			Capabilities: ControllerCapabilities{
				NumPads:      16,
				NumKnobs:     8,
				HasDisplay:   true,
				HasTransport: true,
			},
		},

		// Behringer
		`(?i)behringer.*x32`: {
			Manufacturer: "Behringer",
			Type:         ControllerTypeMixer,
			Capabilities: ControllerCapabilities{
				NumFaders:  32,
				NumKnobs:   16,
				HasDisplay: true,
			},
		},
		`(?i)behringer.*xr18`: {
			Manufacturer: "Behringer",
			Type:         ControllerTypeMixer,
			Capabilities: ControllerCapabilities{
				NumFaders:  18,
				NumKnobs:   8,
				HasDisplay: true,
			},
		},

		// Arturia
		`(?i)arturia.*keylab`: {
			Manufacturer: "Arturia",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      61,
				NumKnobs:     9,
				NumSliders:   9,
				HasTransport: true,
			},
		},
		`(?i)arturia.*beatstep`: {
			Manufacturer: "Arturia",
			Type:         ControllerTypeDrumMachine,
			Capabilities: ControllerCapabilities{
				NumPads:      16,
				NumKnobs:     16,
				HasTransport: true,
			},
		},

		// Novation
		`(?i)novation.*launchkey`: {
			Manufacturer: "Novation",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      25,
				NumKnobs:     8,
				NumPads:      16,
				HasTransport: true,
			},
		},
		`(?i)novation.*launchpad`: {
			Manufacturer: "Novation",
			Type:         ControllerTypePad,
			Capabilities: ControllerCapabilities{
				NumPads:    64,
				HasDisplay: true,
			},
		},

		// M-Audio
		`(?i)m-audio.*oxygen`: {
			Manufacturer: "M-Audio",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      49,
				NumKnobs:     8,
				NumSliders:   9,
				HasTransport: true,
			},
		},

		// Korg
		`(?i)korg.*nano`: {
			Manufacturer: "Korg",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      25,
				NumKnobs:     8,
				NumSliders:   9,
				HasTransport: true,
			},
		},

		// Roland
		`(?i)roland.*a-`: {
			Manufacturer: "Roland",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      49,
				HasModWheel:  true,
				HasPitchBend: true,
			},
		},

		// Yamaha
		`(?i)yamaha.*motif`: {
			Manufacturer: "Yamaha",
			Type:         ControllerTypeSynthesizer,
			Capabilities: ControllerCapabilities{
				NumKeys:      88,
				NumKnobs:     8,
				HasModWheel:  true,
				HasPitchBend: true,
				HasDisplay:   true,
			},
		},

		// Generic Patterns
		`(?i)midi.*keyboard`: {
			Manufacturer: "Unknown",
			Type:         ControllerTypeKeyboard,
			Capabilities: ControllerCapabilities{
				NumKeys:      49,
				HasModWheel:  true,
				HasPitchBend: true,
			},
		},
		`(?i)midi.*controller`: {
			Manufacturer: "Unknown",
			Type:         ControllerTypeKnob,
			Capabilities: ControllerCapabilities{
				NumKnobs:   8,
				NumSliders: 8,
			},
		},
	}

	// Port-Namen gegen Patterns testen
	for pattern, template := range patterns {
		matched, err := regexp.MatchString(pattern, portName)
		if err != nil {
			continue
		}

		if matched {
			// Controller-Info erstellen
			controller := &ControllerInfo{
				ID:           dm.generateControllerID(portName),
				Name:         dm.extractControllerName(portName),
				Manufacturer: template.Manufacturer,
				Model:        dm.extractModelName(portName),
				Type:         template.Type,
				Capabilities: template.Capabilities,
				Settings:     dm.getDefaultSettings(template.Type),
			}

			// Spezifische Anpassungen basierend auf Modell
			dm.adjustCapabilities(controller, portName)

			return controller
		}
	}

	return nil
}

// testPort testet einen MIDI-Port auf Funktionalität
func (dm *DiscoveryManager) testPort(portName string, controller *ControllerInfo) error {
	// TODO: Implementierung des Port-Tests
	// - Port öffnen
	// - Test-Events senden
	// - Antworten überwachen
	// - Capabilities verifizieren

	dm.logger.Debug("Port-Test erfolgreich", "port", portName)
	return nil
}

// generateControllerID generiert eine eindeutige ID für den Controller
func (dm *DiscoveryManager) generateControllerID(portName string) string {
	// Einfache ID-Generierung basierend auf Port-Namen
	cleanName := strings.ReplaceAll(portName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "-", "_")
	cleanName = strings.ToLower(cleanName)
	return cleanName
}

// extractControllerName extrahiert den Controller-Namen aus dem Port-Namen
func (dm *DiscoveryManager) extractControllerName(portName string) string {
	// Entferne häufige Präfixe/Suffixe
	name := portName
	name = strings.TrimPrefix(name, "MIDI ")
	name = strings.TrimPrefix(name, "USB ")
	name = strings.TrimSuffix(name, " MIDI")
	name = strings.TrimSuffix(name, " USB")
	return name
}

// extractModelName extrahiert den Modell-Namen aus dem Port-Namen
func (dm *DiscoveryManager) extractModelName(portName string) string {
	// Suche nach Modell-Nummern oder spezifischen Namen
	modelPatterns := []string{
		`(?i)(mpk\s*\d+)`,
		`(?i)(mpx\s*\d+)`,
		`(?i)(x32)`,
		`(?i)(xr18)`,
		`(?i)(keylab\s*\d+)`,
		`(?i)(beatstep\s*\w*)`,
		`(?i)(launchkey\s*\d+)`,
		`(?i)(launchpad\s*\w*)`,
		`(?i)(oxygen\s*\d+)`,
		`(?i)(nano\w*)`,
		`(?i)(a-\d+)`,
		`(?i)(motif\s*\w*)`,
	}

	for _, pattern := range modelPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(portName)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return "Unknown Model"
}

// getDefaultSettings gibt Standard-Einstellungen für einen Controller-Typ zurück
func (dm *DiscoveryManager) getDefaultSettings(controllerType ControllerType) ControllerSettings {
	settings := ControllerSettings{
		DefaultChannel:    0,
		VelocitySensitive: true,
		PressureSensitive: false,
		AutoMap:           true,
		CustomMappings:    make(map[string]string),
	}

	switch controllerType {
	case ControllerTypeKeyboard:
		settings.VelocitySensitive = true
		settings.PressureSensitive = true
	case ControllerTypePad:
		settings.VelocitySensitive = true
		settings.PressureSensitive = true
	case ControllerTypeKnob:
		settings.VelocitySensitive = false
		settings.PressureSensitive = false
	case ControllerTypeSlider:
		settings.VelocitySensitive = false
		settings.PressureSensitive = false
	case ControllerTypeMixer:
		settings.VelocitySensitive = false
		settings.PressureSensitive = false
		settings.AutoMap = false
	}

	return settings
}

// adjustCapabilities passt die Capabilities basierend auf spezifischen Modellen an
func (dm *DiscoveryManager) adjustCapabilities(controller *ControllerInfo, portName string) {
	// Spezifische Anpassungen für bekannte Modelle
	switch {
	case strings.Contains(strings.ToLower(portName), "mpk mini"):
		controller.Capabilities.NumKeys = 25
		controller.Capabilities.NumKnobs = 8
		controller.Capabilities.NumPads = 8
	case strings.Contains(strings.ToLower(portName), "mpk249"):
		controller.Capabilities.NumKeys = 49
		controller.Capabilities.NumKnobs = 8
		controller.Capabilities.NumPads = 16
		controller.Capabilities.NumSliders = 8
	case strings.Contains(strings.ToLower(portName), "launchkey mini"):
		controller.Capabilities.NumKeys = 25
		controller.Capabilities.NumKnobs = 8
		controller.Capabilities.NumPads = 16
	case strings.Contains(strings.ToLower(portName), "launchkey 49"):
		controller.Capabilities.NumKeys = 49
		controller.Capabilities.NumKnobs = 8
		controller.Capabilities.NumPads = 16
	case strings.Contains(strings.ToLower(portName), "launchkey 61"):
		controller.Capabilities.NumKeys = 61
		controller.Capabilities.NumKnobs = 8
		controller.Capabilities.NumPads = 16
	}
}

// GetControllerByID gibt einen Controller anhand seiner ID zurück
func (dm *DiscoveryManager) GetControllerByID(id string) (*ControllerInfo, bool) {
	controller, exists := dm.knownControllers[id]
	return controller, exists
}

// GetAllControllers gibt alle bekannten Controller zurück
func (dm *DiscoveryManager) GetAllControllers() []*ControllerInfo {
	controllers := make([]*ControllerInfo, 0, len(dm.knownControllers))
	for _, controller := range dm.knownControllers {
		controllers = append(controllers, controller)
	}
	return controllers
}

// UpdateControllerSettings aktualisiert die Einstellungen eines Controllers
func (dm *DiscoveryManager) UpdateControllerSettings(id string, settings ControllerSettings) error {
	controller, exists := dm.knownControllers[id]
	if !exists {
		return fmt.Errorf("controller mit ID '%s' nicht gefunden", id)
	}

	controller.Settings = settings
	dm.logger.Info("Controller-Einstellungen aktualisiert", "id", id)
	return nil
}

// GetSuggestedMappings gibt vorgeschlagene Mappings für einen Controller zurück
func (dm *DiscoveryManager) GetSuggestedMappings(controllerID string) ([]SuggestedMapping, error) {
	controller, exists := dm.knownControllers[controllerID]
	if !exists {
		return nil, fmt.Errorf("controller mit ID '%s' nicht gefunden", controllerID)
	}

	var suggestions []SuggestedMapping

	switch controller.Type {
	case ControllerTypeKeyboard:
		suggestions = dm.getKeyboardMappings(controller)
	case ControllerTypePad:
		suggestions = dm.getPadMappings(controller)
	case ControllerTypeKnob:
		suggestions = dm.getKnobMappings(controller)
	case ControllerTypeSlider:
		suggestions = dm.getSliderMappings(controller)
	case ControllerTypeMixer:
		suggestions = dm.getMixerMappings(controller)
	case ControllerTypeDJ:
		suggestions = dm.getDJMappings(controller)
	}

	return suggestions, nil
}

// SuggestedMapping repräsentiert ein vorgeschlagenes Mapping
type SuggestedMapping struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Event       MIDIEvent       `json:"event"`
	Action      SuggestedAction `json:"action"`
	Priority    int             `json:"priority"` // Höhere Zahl = höhere Priorität
	Category    string          `json:"category"`
}

// SuggestedAction repräsentiert eine vorgeschlagene Aktion
type SuggestedAction struct {
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description"`
}

// getKeyboardMappings gibt vorgeschlagene Mappings für Keyboards zurück
func (dm *DiscoveryManager) getKeyboardMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Volume Control",
			Description: "Lautstärke über Modulationsrad steuern",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 1, // Mod Wheel
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "set",
					"volume":    "{{value}}",
				},
				Description: "Lautstärke auf Controller-Wert setzen",
			},
			Priority: 10,
			Category: "Volume",
		},
		{
			Name:        "Mute Toggle",
			Description: "Stummschaltung über Sustain-Pedal",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 64, // Sustain Pedal
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "mute",
				},
				Description: "Stummschaltung umschalten",
			},
			Priority: 8,
			Category: "Volume",
		},
	}
}

// getPadMappings gibt vorgeschlagene Mappings für Pads zurück
func (dm *DiscoveryManager) getPadMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Screenshot",
			Description: "Screenshot mit Pad 1",
			Event: MIDIEvent{
				Type:     "note_on",
				Note:     36, // Pad 1
				Velocity: 100,
			},
			Action: SuggestedAction{
				Type: "key_combination",
				Parameters: map[string]interface{}{
					"keys": []string{"CTRL", "SHIFT", "S"},
				},
				Description: "Screenshot-Tastenkombination",
			},
			Priority: 9,
			Category: "System",
		},
		{
			Name:        "Copy",
			Description: "Kopieren mit Pad 2",
			Event: MIDIEvent{
				Type:     "note_on",
				Note:     37, // Pad 2
				Velocity: 100,
			},
			Action: SuggestedAction{
				Type: "key_combination",
				Parameters: map[string]interface{}{
					"keys": []string{"CTRL", "C"},
				},
				Description: "Kopieren-Tastenkombination",
			},
			Priority: 8,
			Category: "System",
		},
	}
}

// getKnobMappings gibt vorgeschlagene Mappings für Drehregler zurück
func (dm *DiscoveryManager) getKnobMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Volume Knob 1",
			Description: "Lautstärke über ersten Drehregler",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 1,
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "set",
					"volume":    "{{value}}",
				},
				Description: "Lautstärke auf Controller-Wert setzen",
			},
			Priority: 10,
			Category: "Volume",
		},
	}
}

// getSliderMappings gibt vorgeschlagene Mappings für Schieberegler zurück
func (dm *DiscoveryManager) getSliderMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Volume Slider 1",
			Description: "Lautstärke über ersten Schieberegler",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 7, // Volume
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "set",
					"volume":    "{{value}}",
				},
				Description: "Lautstärke auf Controller-Wert setzen",
			},
			Priority: 10,
			Category: "Volume",
		},
	}
}

// getMixerMappings gibt vorgeschlagene Mappings für Mixer zurück
func (dm *DiscoveryManager) getMixerMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Master Volume",
			Description: "Master-Lautstärke über Hauptfader",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 7, // Volume
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "set",
					"volume":    "{{value}}",
				},
				Description: "System-Lautstärke auf Fader-Wert setzen",
			},
			Priority: 10,
			Category: "Volume",
		},
	}
}

// getDJMappings gibt vorgeschlagene Mappings für DJ-Controller zurück
func (dm *DiscoveryManager) getDJMappings(controller *ControllerInfo) []SuggestedMapping {
	return []SuggestedMapping{
		{
			Name:        "Play/Pause",
			Description: "Play/Pause über Transport-Button",
			Event: MIDIEvent{
				Type:     "note_on",
				Note:     118, // Play
				Velocity: 100,
			},
			Action: SuggestedAction{
				Type: "key_combination",
				Parameters: map[string]interface{}{
					"keys": []string{"SPACE"},
				},
				Description: "Leertaste für Play/Pause",
			},
			Priority: 9,
			Category: "Transport",
		},
		{
			Name:        "Volume Control",
			Description: "Lautstärke über Hauptfader",
			Event: MIDIEvent{
				Type:       "control_change",
				Controller: 7, // Volume
			},
			Action: SuggestedAction{
				Type: "volume",
				Parameters: map[string]interface{}{
					"direction": "set",
					"volume":    "{{value}}",
				},
				Description: "System-Lautstärke auf Fader-Wert setzen",
			},
			Priority: 10,
			Category: "Volume",
		},
	}
}
