// Package actions verwaltet die Ausführung von Systemaktionen basierend auf MIDI-Events.
// Diese Datei enthält den Audio-Source-Executor für das Wechseln von Audioquellen.

package actions

import (
	"fmt"
	"runtime"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// AudioSourceExecutor verwaltet das Wechseln von Audioquellen
type AudioSourceExecutor struct {
	BaseExecutor
	audioController AudioController
}

// AudioController definiert die Schnittstelle für plattformspezifische Audioquellen-Steuerung
type AudioController interface {
	GetAudioSources() ([]AudioSource, error)
	SetDefaultAudioSource(sourceID string) error
	GetDefaultAudioSource() (AudioSource, error)
	MuteAudioSource(sourceID string) error
	UnmuteAudioSource(sourceID string) error
	SetAudioSourceVolume(sourceID string, volume int) error
}

// AudioSource repräsentiert eine Audioquelle
type AudioSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "speakers", "headphones", "hdmi", "bluetooth", etc.
	IsDefault   bool   `json:"is_default"`
	IsMuted     bool   `json:"is_muted"`
	Volume      int    `json:"volume"`
	IsAvailable bool   `json:"is_available"`
}

// NewAudioSourceExecutor erstellt einen neuen Audio-Source-Executor
func NewAudioSourceExecutor(logger utils.Logger) (*AudioSourceExecutor, error) {
	// Plattformspezifischen Audio-Controller erstellen
	controller, err := newAudioController()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Audio-Controllers: %w", err)
	}

	executor := &AudioSourceExecutor{
		BaseExecutor:    NewBaseExecutor("audio_source", logger),
		audioController: controller,
	}

	return executor, nil
}

// Execute führt eine Audio-Source-Aktion aus
func (e *AudioSourceExecutor) Execute(action config.Action) error {
	e.LogDebug("Führe Audio-Source-Aktion aus", "parameters", action.Parameters)

	// Source-Parameter extrahieren
	source, ok := action.Parameters["source"]
	if !ok {
		return fmt.Errorf("audio_source-Aktion benötigt 'source' Parameter")
	}

	sourceStr, ok := source.(string)
	if !ok {
		return fmt.Errorf("'source' Parameter muss ein String sein")
	}

	// Aktionstyp bestimmen
	actionType := "switch"
	if typeParam, ok := action.Parameters["type"]; ok {
		if typeStr, ok := typeParam.(string); ok {
			actionType = typeStr
		}
	}

	// Audio-Source-Aktion ausführen
	switch actionType {
	case "switch":
		e.LogInfo("Wechsle Audioquelle", "source", sourceStr)
		return e.audioController.SetDefaultAudioSource(sourceStr)

	case "mute":
		e.LogInfo("Stummschalten Audioquelle", "source", sourceStr)
		return e.audioController.MuteAudioSource(sourceStr)

	case "unmute":
		e.LogInfo("Stummschaltung aufheben Audioquelle", "source", sourceStr)
		return e.audioController.UnmuteAudioSource(sourceStr)

	case "volume":
		// Lautstärke setzen
		volume := 50 // Standard: 50%
		if volumeParam, ok := action.Parameters["volume"]; ok {
			switch v := volumeParam.(type) {
			case int:
				volume = v
			case float64:
				volume = int(v)
			case string:
				if parsed, err := parseInt(v); err == nil {
					volume = parsed
				}
			}
		}

		if volume < 0 || volume > 100 {
			return fmt.Errorf("volume muss zwischen 0 und 100 liegen, got: %d", volume)
		}

		e.LogInfo("Setze Audioquelle-Lautstärke", "source", sourceStr, "volume", volume)
		return e.audioController.SetAudioSourceVolume(sourceStr, volume)

	case "cycle":
		// Durch verfügbare Quellen wechseln
		e.LogInfo("Wechsle zur nächsten Audioquelle")
		return e.cycleAudioSource()

	default:
		return fmt.Errorf("ungültiger Aktionstyp: %s (erwartet: switch, mute, unmute, volume, cycle)", actionType)
	}
}

// cycleAudioSource wechselt zur nächsten verfügbaren Audioquelle
func (e *AudioSourceExecutor) cycleAudioSource() error {
	sources, err := e.audioController.GetAudioSources()
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der Audioquellen: %w", err)
	}

	if len(sources) == 0 {
		return fmt.Errorf("keine Audioquellen verfügbar")
	}

	// Aktuelle Standardquelle finden
	currentSource, err := e.audioController.GetDefaultAudioSource()
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der aktuellen Audioquelle: %w", err)
	}

	// Nächste verfügbare Quelle finden
	var nextSource *AudioSource
	for i, source := range sources {
		if source.ID == currentSource.ID {
			// Nächste Quelle in der Liste
			nextIndex := (i + 1) % len(sources)
			nextSource = &sources[nextIndex]
			break
		}
	}

	if nextSource == nil {
		// Fallback: erste verfügbare Quelle
		nextSource = &sources[0]
	}

	e.LogInfo("Wechsle zu Audioquelle", "from", currentSource.Name, "to", nextSource.Name)
	return e.audioController.SetDefaultAudioSource(nextSource.ID)
}

// Validate überprüft eine Audio-Source-Aktion auf Gültigkeit
func (e *AudioSourceExecutor) Validate(action config.Action) error {
	// Source-Parameter überprüfen
	source, ok := action.Parameters["source"]
	if !ok {
		return fmt.Errorf("audio_source-Aktion benötigt 'source' Parameter")
	}

	sourceStr, ok := source.(string)
	if !ok {
		return fmt.Errorf("'source' Parameter muss ein String sein")
	}

	if sourceStr == "" {
		return fmt.Errorf("'source' Parameter darf nicht leer sein")
	}

	// Type-Parameter überprüfen (falls vorhanden)
	if typeParam, ok := action.Parameters["type"]; ok {
		if typeStr, ok := typeParam.(string); ok {
			validTypes := map[string]bool{
				"switch": true,
				"mute":   true,
				"unmute": true,
				"volume": true,
				"cycle":  true,
			}
			if !validTypes[typeStr] {
				return fmt.Errorf("ungültiger Typ: %s", typeStr)
			}

			// Bei "cycle" Typ ist source-Parameter optional
			if typeStr == "cycle" {
				// Source-Parameter wird ignoriert
			} else {
				// Source-Validierung (falls möglich)
				if err := e.validateSource(sourceStr); err != nil {
					return fmt.Errorf("ungültige Audioquelle: %w", err)
				}
			}
		} else {
			return fmt.Errorf("'type' Parameter muss ein String sein")
		}
	}

	// Volume-Parameter überprüfen (falls vorhanden)
	if volumeParam, ok := action.Parameters["volume"]; ok {
		var volume int
		switch v := volumeParam.(type) {
		case int:
			volume = v
		case float64:
			volume = int(v)
		case string:
			if parsed, err := parseInt(v); err != nil {
				return fmt.Errorf("ungültiger 'volume' Parameter: %v", volumeParam)
			} else {
				volume = parsed
			}
		default:
			return fmt.Errorf("ungültiger 'volume' Parameter: %v", volumeParam)
		}

		if volume < 0 || volume > 100 {
			return fmt.Errorf("volume muss zwischen 0 und 100 liegen, got: %d", volume)
		}
	}

	return nil
}

// validateSource überprüft eine Audioquelle auf Gültigkeit
func (e *AudioSourceExecutor) validateSource(sourceID string) error {
	// Verfügbare Quellen abrufen
	sources, err := e.audioController.GetAudioSources()
	if err != nil {
		// Bei Fehlern trotzdem erlauben (könnte ein neues Gerät sein)
		return nil
	}

	// Überprüfen ob die Quelle existiert
	for _, source := range sources {
		if source.ID == sourceID || source.Name == sourceID {
			return nil
		}
	}

	return fmt.Errorf("audioquelle '%s' nicht gefunden", sourceID)
}

// GetAvailableSources gibt alle verfügbaren Audioquellen zurück
func (e *AudioSourceExecutor) GetAvailableSources() ([]AudioSource, error) {
	return e.audioController.GetAudioSources()
}

// GetCurrentSource gibt die aktuelle Standard-Audioquelle zurück
func (e *AudioSourceExecutor) GetCurrentSource() (AudioSource, error) {
	return e.audioController.GetDefaultAudioSource()
}

// newAudioController erstellt einen plattformspezifischen Audio-Controller
func newAudioController() (AudioController, error) {
	switch runtime.GOOS {
	case "windows":
		return newWindowsAudioController()
	case "linux":
		return newLinuxAudioController()
	default:
		return nil, fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// Windows-spezifische Audio-Controller-Implementierung
type windowsAudioController struct{}

func newWindowsAudioController() (AudioController, error) {
	return &windowsAudioController{}, nil
}

func (c *windowsAudioController) GetAudioSources() ([]AudioSource, error) {
	// TODO: Implementierung mit Windows Core Audio API
	// - IMMDeviceEnumerator verwenden
	// - Alle verfügbaren Audio-Endpunkte auflisten
	return []AudioSource{
		{ID: "speakers", Name: "Lautsprecher", Type: "speakers", IsDefault: true, IsMuted: false, Volume: 50, IsAvailable: true},
		{ID: "headphones", Name: "Kopfhörer", Type: "headphones", IsDefault: false, IsMuted: false, Volume: 50, IsAvailable: true},
		{ID: "hdmi", Name: "HDMI Audio", Type: "hdmi", IsDefault: false, IsMuted: false, Volume: 50, IsAvailable: true},
	}, nil
}

func (c *windowsAudioController) SetDefaultAudioSource(sourceID string) error {
	// TODO: Implementierung mit Windows Core Audio API
	// - IMMDeviceEnumerator.SetDefaultEndpoint aufrufen
	return nil
}

func (c *windowsAudioController) GetDefaultAudioSource() (AudioSource, error) {
	// TODO: Implementierung mit Windows Core Audio API
	// - IMMDeviceEnumerator.GetDefaultAudioEndpoint aufrufen
	return AudioSource{ID: "speakers", Name: "Lautsprecher", Type: "speakers", IsDefault: true, IsMuted: false, Volume: 50, IsAvailable: true}, nil
}

func (c *windowsAudioController) MuteAudioSource(sourceID string) error {
	// TODO: Implementierung mit Windows Core Audio API
	// - IAudioEndpointVolume.SetMute aufrufen
	return nil
}

func (c *windowsAudioController) UnmuteAudioSource(sourceID string) error {
	// TODO: Implementierung mit Windows Core Audio API
	// - IAudioEndpointVolume.SetMute aufrufen
	return nil
}

func (c *windowsAudioController) SetAudioSourceVolume(sourceID string, volume int) error {
	// TODO: Implementierung mit Windows Core Audio API
	// - IAudioEndpointVolume.SetMasterVolumeLevelScalar aufrufen
	return nil
}

// Linux-spezifische Audio-Controller-Implementierung
type linuxAudioController struct{}

func newLinuxAudioController() (AudioController, error) {
	return &linuxAudioController{}, nil
}

func (c *linuxAudioController) GetAudioSources() ([]AudioSource, error) {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl list sinks verwenden (PulseAudio)
	// - oder amixer -c 0 scontrols verwenden (ALSA)
	return []AudioSource{
		{ID: "analog-stereo", Name: "Analoger Stereo-Ausgang", Type: "speakers", IsDefault: true, IsMuted: false, Volume: 50, IsAvailable: true},
		{ID: "hdmi-stereo", Name: "HDMI Stereo", Type: "hdmi", IsDefault: false, IsMuted: false, Volume: 50, IsAvailable: true},
		{ID: "bluetooth", Name: "Bluetooth Audio", Type: "bluetooth", IsDefault: false, IsMuted: false, Volume: 50, IsAvailable: true},
	}, nil
}

func (c *linuxAudioController) SetDefaultAudioSource(sourceID string) error {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl set-default-sink verwenden (PulseAudio)
	// - oder amixer -c 0 sset verwenden (ALSA)
	return nil
}

func (c *linuxAudioController) GetDefaultAudioSource() (AudioSource, error) {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl info verwenden (PulseAudio)
	// - oder amixer -c 0 sget verwenden (ALSA)
	return AudioSource{ID: "analog-stereo", Name: "Analoger Stereo-Ausgang", Type: "speakers", IsDefault: true, IsMuted: false, Volume: 50, IsAvailable: true}, nil
}

func (c *linuxAudioController) MuteAudioSource(sourceID string) error {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl set-sink-mute verwenden (PulseAudio)
	// - oder amixer -c 0 sset verwenden (ALSA)
	return nil
}

func (c *linuxAudioController) UnmuteAudioSource(sourceID string) error {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl set-sink-mute verwenden (PulseAudio)
	// - oder amixer -c 0 sset verwenden (ALSA)
	return nil
}

func (c *linuxAudioController) SetAudioSourceVolume(sourceID string, volume int) error {
	// TODO: Implementierung mit PulseAudio oder ALSA
	// - pactl set-sink-volume verwenden (PulseAudio)
	// - oder amixer -c 0 sset verwenden (ALSA)
	return nil
}

// parseInt ist eine Hilfsfunktion zum Parsen von Strings zu Integers
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
