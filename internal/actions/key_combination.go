// Package actions verwaltet die Ausführung von Systemaktionen basierend auf MIDI-Events.
// Diese Datei enthält den Key-Combination-Executor für das Senden von Tastenkombinationen.

package actions

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// KeyCombinationExecutor verwaltet das Senden von Tastenkombinationen
type KeyCombinationExecutor struct {
	BaseExecutor
	keyboard KeyboardController
}

// KeyboardController definiert die Schnittstelle für plattformspezifische Tastatureingaben
type KeyboardController interface {
	SendKey(key string) error
	SendKeyCombination(keys []string) error
	SendText(text string) error
	HoldKey(key string, duration time.Duration) error
}

// NewKeyCombinationExecutor erstellt einen neuen Key-Combination-Executor
func NewKeyCombinationExecutor(logger utils.Logger) (*KeyCombinationExecutor, error) {
	// Plattformspezifischen Keyboard-Controller erstellen
	controller, err := newKeyboardController()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Keyboard-Controllers: %w", err)
	}

	executor := &KeyCombinationExecutor{
		BaseExecutor: NewBaseExecutor("key_combination", logger),
		keyboard:     controller,
	}

	return executor, nil
}

// Execute führt eine Key-Combination-Aktion aus
func (e *KeyCombinationExecutor) Execute(action config.Action) error {
	e.LogDebug("Führe Key-Combination-Aktion aus", "parameters", action.Parameters)

	// Keys-Parameter extrahieren
	keys, ok := action.Parameters["keys"]
	if !ok {
		return fmt.Errorf("key_combination-Aktion benötigt 'keys' Parameter")
	}

	var keyList []string
	switch v := keys.(type) {
	case []interface{}:
		for _, key := range v {
			if keyStr, ok := key.(string); ok {
				keyList = append(keyList, keyStr)
			}
		}
	case []string:
		keyList = v
	case string:
		// Komma-getrennte Liste
		keyList = strings.Split(v, ",")
		for i, key := range keyList {
			keyList[i] = strings.TrimSpace(key)
		}
	default:
		return fmt.Errorf("ungültiger 'keys' Parameter: %v", keys)
	}

	if len(keyList) == 0 {
		return fmt.Errorf("'keys' Parameter darf nicht leer sein")
	}

	// Verzögerung zwischen Tasten extrahieren (optional)
	delay := 50 * time.Millisecond // Standard: 50ms
	if delayParam, ok := action.Parameters["delay"]; ok {
		switch v := delayParam.(type) {
		case int:
			delay = time.Duration(v) * time.Millisecond
		case float64:
			delay = time.Duration(v) * time.Millisecond
		case string:
			if parsed, err := time.ParseDuration(v); err == nil {
				delay = parsed
			}
		}
	}

	// Aktionstyp bestimmen
	actionType := "combination"
	if typeParam, ok := action.Parameters["type"]; ok {
		if typeStr, ok := typeParam.(string); ok {
			actionType = typeStr
		}
	}

	// Tastenkombination ausführen
	switch actionType {
	case "combination":
		e.LogInfo("Sende Tastenkombination", "keys", keyList, "delay", delay)
		return e.keyboard.SendKeyCombination(keyList)

	case "sequence":
		e.LogInfo("Sende Tastensequenz", "keys", keyList, "delay", delay)
		for i, key := range keyList {
			if err := e.keyboard.SendKey(key); err != nil {
				return fmt.Errorf("fehler beim Senden der Taste '%s': %w", key, err)
			}
			if i < len(keyList)-1 && delay > 0 {
				time.Sleep(delay)
			}
		}
		return nil

	case "hold":
		// Taste gedrückt halten
		if len(keyList) != 1 {
			return fmt.Errorf("'hold' Typ benötigt genau eine Taste")
		}
		holdDuration := 1 * time.Second // Standard: 1 Sekunde
		if durationParam, ok := action.Parameters["duration"]; ok {
			switch v := durationParam.(type) {
			case int:
				holdDuration = time.Duration(v) * time.Millisecond
			case float64:
				holdDuration = time.Duration(v) * time.Millisecond
			case string:
				if parsed, err := time.ParseDuration(v); err == nil {
					holdDuration = parsed
				}
			}
		}
		e.LogInfo("Halte Taste gedrückt", "key", keyList[0], "duration", holdDuration)
		return e.keyboard.HoldKey(keyList[0], holdDuration)

	case "text":
		// Text eingeben
		text := strings.Join(keyList, "")
		e.LogInfo("Gebe Text ein", "text", text)
		return e.keyboard.SendText(text)

	default:
		return fmt.Errorf("ungültiger Aktionstyp: %s (erwartet: combination, sequence, hold, text)", actionType)
	}
}

// Validate überprüft eine Key-Combination-Aktion auf Gültigkeit
func (e *KeyCombinationExecutor) Validate(action config.Action) error {
	// Keys-Parameter überprüfen
	keys, ok := action.Parameters["keys"]
	if !ok {
		return fmt.Errorf("key_combination-Aktion benötigt 'keys' Parameter")
	}

	var keyList []string
	switch v := keys.(type) {
	case []interface{}:
		for i, key := range v {
			if keyStr, ok := key.(string); ok {
				keyList = append(keyList, keyStr)
			} else {
				return fmt.Errorf("key %d muss ein String sein", i)
			}
		}
	case []string:
		keyList = v
	case string:
		// OK
	default:
		return fmt.Errorf("ungültiger 'keys' Parameter: %v", keys)
	}

	if len(keyList) == 0 {
		return fmt.Errorf("'keys' Parameter darf nicht leer sein")
	}

	// Tasten validieren
	for i, key := range keyList {
		if err := e.validateKey(key); err != nil {
			return fmt.Errorf("ungültige Taste %d: %w", i, err)
		}
	}

	// Type-Parameter überprüfen (falls vorhanden)
	if typeParam, ok := action.Parameters["type"]; ok {
		if typeStr, ok := typeParam.(string); ok {
			validTypes := map[string]bool{
				"combination": true,
				"sequence":    true,
				"hold":        true,
				"text":        true,
			}
			if !validTypes[typeStr] {
				return fmt.Errorf("ungültiger Typ: %s", typeStr)
			}

			// Spezielle Validierung für "hold" Typ
			if typeStr == "hold" && len(keyList) != 1 {
				return fmt.Errorf("'hold' Typ benötigt genau eine Taste")
			}
		} else {
			return fmt.Errorf("'type' Parameter muss ein String sein")
		}
	}

	// Delay-Parameter überprüfen (falls vorhanden)
	if delayParam, ok := action.Parameters["delay"]; ok {
		switch v := delayParam.(type) {
		case int:
			if v < 0 {
				return fmt.Errorf("delay muss positiv sein")
			}
		case float64:
			if v < 0 {
				return fmt.Errorf("delay muss positiv sein")
			}
		case string:
			if _, err := time.ParseDuration(v); err != nil {
				return fmt.Errorf("ungültiger delay: %s", v)
			}
		default:
			return fmt.Errorf("ungültiger 'delay' Parameter: %v", delayParam)
		}
	}

	// Duration-Parameter überprüfen (falls vorhanden)
	if durationParam, ok := action.Parameters["duration"]; ok {
		switch v := durationParam.(type) {
		case int:
			if v < 0 {
				return fmt.Errorf("duration muss positiv sein")
			}
		case float64:
			if v < 0 {
				return fmt.Errorf("duration muss positiv sein")
			}
		case string:
			if _, err := time.ParseDuration(v); err != nil {
				return fmt.Errorf("ungültige duration: %s", v)
			}
		default:
			return fmt.Errorf("ungültiger 'duration' Parameter: %v", durationParam)
		}
	}

	return nil
}

// validateKey überprüft eine einzelne Taste auf Gültigkeit
func (e *KeyCombinationExecutor) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("taste darf nicht leer sein")
	}

	// Plattformspezifische Tastenvalidierung
	switch runtime.GOOS {
	case "windows":
		return e.validateWindowsKey(key)
	case "linux":
		return e.validateLinuxKey(key)
	default:
		return fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// validateWindowsKey überprüft eine Windows-Taste
func (e *KeyCombinationExecutor) validateWindowsKey(key string) error {
	// Gültige Windows-Tasten
	validKeys := map[string]bool{
		// Funktionstasten
		"F1": true, "F2": true, "F3": true, "F4": true, "F5": true, "F6": true,
		"F7": true, "F8": true, "F9": true, "F10": true, "F11": true, "F12": true,
		// Modifier-Tasten
		"CTRL": true, "ALT": true, "SHIFT": true, "WIN": true,
		// Navigation
		"UP": true, "DOWN": true, "LEFT": true, "RIGHT": true,
		"HOME": true, "END": true, "PAGEUP": true, "PAGEDOWN": true,
		// Andere
		"ENTER": true, "ESC": true, "TAB": true, "SPACE": true,
		"BACKSPACE": true, "DELETE": true, "INSERT": true,
		// Buchstaben und Zahlen sind auch gültig
	}

	// Einzelne Buchstaben und Zahlen sind gültig
	if len(key) == 1 {
		return nil
	}

	// Spezielle Tasten überprüfen
	if !validKeys[strings.ToUpper(key)] {
		return fmt.Errorf("ungültige Taste: %s", key)
	}

	return nil
}

// validateLinuxKey überprüft eine Linux-Taste
func (e *KeyCombinationExecutor) validateLinuxKey(key string) error {
	// Gültige Linux-Tasten (X11)
	validKeys := map[string]bool{
		// Funktionstasten
		"F1": true, "F2": true, "F3": true, "F4": true, "F5": true, "F6": true,
		"F7": true, "F8": true, "F9": true, "F10": true, "F11": true, "F12": true,
		// Modifier-Tasten
		"CTRL": true, "ALT": true, "SHIFT": true, "SUPER": true,
		// Navigation
		"UP": true, "DOWN": true, "LEFT": true, "RIGHT": true,
		"HOME": true, "END": true, "PAGEUP": true, "PAGEDOWN": true,
		// Andere
		"ENTER": true, "ESC": true, "TAB": true, "SPACE": true,
		"BACKSPACE": true, "DELETE": true, "INSERT": true,
		// Buchstaben und Zahlen sind auch gültig
	}

	// Einzelne Buchstaben und Zahlen sind gültig
	if len(key) == 1 {
		return nil
	}

	// Spezielle Tasten überprüfen
	if !validKeys[strings.ToUpper(key)] {
		return fmt.Errorf("ungültige Taste: %s", key)
	}

	return nil
}

// newKeyboardController erstellt einen plattformspezifischen Keyboard-Controller
func newKeyboardController() (KeyboardController, error) {
	switch runtime.GOOS {
	case "windows":
		return newWindowsKeyboardController()
	case "linux":
		return newLinuxKeyboardController()
	default:
		return nil, fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// Windows-spezifische Keyboard-Controller-Implementierung
type windowsKeyboardController struct{}

func newWindowsKeyboardController() (KeyboardController, error) {
	return &windowsKeyboardController{}, nil
}

func (c *windowsKeyboardController) SendKey(key string) error {
	// TODO: Implementierung mit Windows API
	// - keybd_event oder SendInput verwenden
	// - Taste drücken und loslassen
	return nil
}

func (c *windowsKeyboardController) SendKeyCombination(keys []string) error {
	// TODO: Implementierung mit Windows API
	// - Alle Modifier-Tasten drücken
	// - Haupttaste drücken und loslassen
	// - Modifier-Tasten loslassen
	return nil
}

func (c *windowsKeyboardController) SendText(text string) error {
	// TODO: Implementierung mit Windows API
	// - Jedes Zeichen einzeln senden
	return nil
}

func (c *windowsKeyboardController) HoldKey(key string, duration time.Duration) error {
	// TODO: Implementierung mit Windows API
	// - Taste drücken
	// - Warten
	// - Taste loslassen
	return nil
}

// Linux-spezifische Keyboard-Controller-Implementierung
type linuxKeyboardController struct{}

func newLinuxKeyboardController() (KeyboardController, error) {
	return &linuxKeyboardController{}, nil
}

func (c *linuxKeyboardController) SendKey(key string) error {
	// TODO: Implementierung mit X11 oder uinput
	// - XTestFakeKeyEvent verwenden
	return nil
}

func (c *linuxKeyboardController) SendKeyCombination(keys []string) error {
	// TODO: Implementierung mit X11 oder uinput
	// - Alle Tasten in der richtigen Reihenfolge senden
	return nil
}

func (c *linuxKeyboardController) SendText(text string) error {
	// TODO: Implementierung mit X11 oder uinput
	// - Jedes Zeichen einzeln senden
	return nil
}

func (c *linuxKeyboardController) HoldKey(key string, duration time.Duration) error {
	// TODO: Implementierung mit X11 oder uinput
	// - Taste drücken, warten, loslassen
	return nil
} 