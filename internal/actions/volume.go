// Package actions verwaltet die Ausführung von Systemaktionen basierend auf MIDI-Events.
// Diese Datei enthält den Volume-Executor für die Lautstärkesteuerung.

package actions

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// VolumeExecutor verwaltet die Lautstärkesteuerung
type VolumeExecutor struct {
	BaseExecutor
	volumeController VolumeController
}

// VolumeController definiert die Schnittstelle für plattformspezifische Lautstärkesteuerung
type VolumeController interface {
	GetVolume() (int, error)
	SetVolume(volume int) error
	IncreaseVolume(percent int) error
	DecreaseVolume(percent int) error
}

// NewVolumeExecutor erstellt einen neuen Volume-Executor
func NewVolumeExecutor(logger utils.Logger) (*VolumeExecutor, error) {
	// Plattformspezifischen Volume-Controller erstellen
	controller, err := newVolumeController()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Volume-Controllers: %w", err)
	}

	executor := &VolumeExecutor{
		BaseExecutor:     NewBaseExecutor("volume", logger),
		volumeController: controller,
	}

	return executor, nil
}

// Execute führt eine Volume-Aktion aus
func (e *VolumeExecutor) Execute(action config.Action) error {
	e.LogDebug("Führe Volume-Aktion aus", "parameters", action.Parameters)

	// Direction-Parameter extrahieren
	direction, ok := action.Parameters["direction"]
	if !ok {
		return fmt.Errorf("volume-Aktion benötigt 'direction' Parameter")
	}

	directionStr, ok := direction.(string)
	if !ok {
		return fmt.Errorf("'direction' Parameter muss ein String sein")
	}

	// Prozent-Parameter extrahieren (optional, Standard: 5%)
	percent := 5
	if percentParam, ok := action.Parameters["percent"]; ok {
		switch v := percentParam.(type) {
		case int:
			percent = v
		case float64:
			percent = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				percent = parsed
			}
		}
	}

	// Volume-Aktion ausführen
	switch directionStr {
	case "up", "increase":
		e.LogInfo("Erhöhe Lautstärke", "percent", percent)
		return e.volumeController.IncreaseVolume(percent)

	case "down", "decrease":
		e.LogInfo("Verringere Lautstärke", "percent", percent)
		return e.volumeController.DecreaseVolume(percent)

	case "set":
		// Spezifische Lautstärke setzen
		if volumeParam, ok := action.Parameters["volume"]; ok {
			var volume int
			switch v := volumeParam.(type) {
			case int:
				volume = v
			case float64:
				volume = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					volume = parsed
				} else {
					return fmt.Errorf("ungültiger 'volume' Parameter: %v", volumeParam)
				}
			default:
				return fmt.Errorf("ungültiger 'volume' Parameter: %v", volumeParam)
			}

			if volume < 0 || volume > 100 {
				return fmt.Errorf("volume muss zwischen 0 und 100 liegen, got: %d", volume)
			}

			e.LogInfo("Setze Lautstärke", "volume", volume)
			return e.volumeController.SetVolume(volume)
		}
		return fmt.Errorf("'set' direction benötigt 'volume' Parameter")

	case "mute":
		e.LogInfo("Stummschalten")
		return e.volumeController.SetVolume(0)

	case "unmute":
		e.LogInfo("Stummschaltung aufheben")
		// Auf 50% setzen als Standard
		return e.volumeController.SetVolume(50)

	default:
		return fmt.Errorf("ungültige direction: %s (erwartet: up, down, set, mute, unmute)", directionStr)
	}
}

// Validate überprüft eine Volume-Aktion auf Gültigkeit
func (e *VolumeExecutor) Validate(action config.Action) error {
	// Direction-Parameter überprüfen
	direction, ok := action.Parameters["direction"]
	if !ok {
		return fmt.Errorf("volume-Aktion benötigt 'direction' Parameter")
	}

	directionStr, ok := direction.(string)
	if !ok {
		return fmt.Errorf("'direction' Parameter muss ein String sein")
	}

	// Gültige Directions überprüfen
	validDirections := map[string]bool{
		"up":      true,
		"down":    true,
		"increase": true,
		"decrease": true,
		"set":     true,
		"mute":    true,
		"unmute":  true,
	}

	if !validDirections[directionStr] {
		return fmt.Errorf("ungültige direction: %s", directionStr)
	}

	// Bei "set" direction muss volume-Parameter vorhanden sein
	if directionStr == "set" {
		if _, ok := action.Parameters["volume"]; !ok {
			return fmt.Errorf("'set' direction benötigt 'volume' Parameter")
		}
	}

	// Prozent-Parameter überprüfen (falls vorhanden)
	if percentParam, ok := action.Parameters["percent"]; ok {
		var percent int
		switch v := percentParam.(type) {
		case int:
			percent = v
		case float64:
			percent = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err != nil {
				return fmt.Errorf("ungültiger 'percent' Parameter: %v", percentParam)
			} else {
				percent = parsed
			}
		default:
			return fmt.Errorf("ungültiger 'percent' Parameter: %v", percentParam)
		}

		if percent <= 0 || percent > 100 {
			return fmt.Errorf("percent muss zwischen 1 und 100 liegen, got: %d", percent)
		}
	}

	return nil
}

// GetCurrentVolume gibt die aktuelle Lautstärke zurück
func (e *VolumeExecutor) GetCurrentVolume() (int, error) {
	return e.volumeController.GetVolume()
}

// newVolumeController erstellt einen plattformspezifischen Volume-Controller
func newVolumeController() (VolumeController, error) {
	switch runtime.GOOS {
	case "windows":
		return newWindowsVolumeController()
	case "linux":
		return newLinuxVolumeController()
	default:
		return nil, fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// Windows-spezifische Volume-Controller-Implementierung
type windowsVolumeController struct{}

func newWindowsVolumeController() (VolumeController, error) {
	return &windowsVolumeController{}, nil
}

func (c *windowsVolumeController) GetVolume() (int, error) {
	// TODO: Implementierung mit Windows API
	// - GetMasterVolumeLevel aufrufen
	// - Lautstärke in Prozent zurückgeben
	return 50, nil // Platzhalter
}

func (c *windowsVolumeController) SetVolume(volume int) error {
	// TODO: Implementierung mit Windows API
	// - SetMasterVolumeLevel aufrufen
	// - Lautstärke auf den angegebenen Wert setzen
	return nil
}

func (c *windowsVolumeController) IncreaseVolume(percent int) error {
	current, err := c.GetVolume()
	if err != nil {
		return err
	}

	newVolume := current + percent
	if newVolume > 100 {
		newVolume = 100
	}

	return c.SetVolume(newVolume)
}

func (c *windowsVolumeController) DecreaseVolume(percent int) error {
	current, err := c.GetVolume()
	if err != nil {
		return err
	}

	newVolume := current - percent
	if newVolume < 0 {
		newVolume = 0
	}

	return c.SetVolume(newVolume)
}

// Linux-spezifische Volume-Controller-Implementierung
type linuxVolumeController struct{}

func newLinuxVolumeController() (VolumeController, error) {
	return &linuxVolumeController{}, nil
}

func (c *linuxVolumeController) GetVolume() (int, error) {
	// TODO: Implementierung mit ALSA oder PulseAudio
	// - amixer get Master aufrufen
	// - Lautstärke aus der Ausgabe extrahieren
	return 50, nil // Platzhalter
}

func (c *linuxVolumeController) SetVolume(volume int) error {
	// TODO: Implementierung mit ALSA oder PulseAudio
	// - amixer set Master aufrufen
	// - Lautstärke auf den angegebenen Wert setzen
	return nil
}

func (c *linuxVolumeController) IncreaseVolume(percent int) error {
	current, err := c.GetVolume()
	if err != nil {
		return err
	}

	newVolume := current + percent
	if newVolume > 100 {
		newVolume = 100
	}

	return c.SetVolume(newVolume)
}

func (c *linuxVolumeController) DecreaseVolume(percent int) error {
	current, err := c.GetVolume()
	if err != nil {
		return err
	}

	newVolume := current - percent
	if newVolume < 0 {
		newVolume = 0
	}

	return c.SetVolume(newVolume)
} 