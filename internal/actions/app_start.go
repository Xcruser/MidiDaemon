// Package actions verwaltet die Ausführung von Systemaktionen basierend auf MIDI-Events.
// Diese Datei enthält den App-Start-Executor für das Starten von Anwendungen.

package actions

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// AppStartExecutor verwaltet das Starten von Anwendungen
type AppStartExecutor struct {
	BaseExecutor
}

// NewAppStartExecutor erstellt einen neuen App-Start-Executor
func NewAppStartExecutor(logger utils.Logger) (*AppStartExecutor, error) {
	executor := &AppStartExecutor{
		BaseExecutor: NewBaseExecutor("app_start", logger),
	}

	return executor, nil
}

// Execute führt eine App-Start-Aktion aus
func (e *AppStartExecutor) Execute(action config.Action) error {
	e.LogDebug("Führe App-Start-Aktion aus", "parameters", action.Parameters)

	// Path-Parameter extrahieren
	path, ok := action.Parameters["path"]
	if !ok {
		return fmt.Errorf("app_start-Aktion benötigt 'path' Parameter")
	}

	pathStr, ok := path.(string)
	if !ok {
		return fmt.Errorf("'path' Parameter muss ein String sein")
	}

	// Argumente extrahieren (optional)
	var args []string
	if argsParam, ok := action.Parameters["args"]; ok {
		switch v := argsParam.(type) {
		case []interface{}:
			for _, arg := range v {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		case []string:
			args = v
		case string:
			// Komma-getrennte Liste
			args = strings.Split(v, ",")
			for i, arg := range args {
				args[i] = strings.TrimSpace(arg)
			}
		}
	}

	// Arbeitsverzeichnis extrahieren (optional)
	var workingDir string
	if dirParam, ok := action.Parameters["working_dir"]; ok {
		if dirStr, ok := dirParam.(string); ok {
			workingDir = dirStr
		}
	}

	// Anwendung starten
	e.LogInfo("Starte Anwendung", "path", pathStr, "args", args, "working_dir", workingDir)

	cmd := exec.Command(pathStr, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Hintergrund ausführen (nicht auf Fertigstellung warten)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("fehler beim Starten der Anwendung '%s': %w", pathStr, err)
	}

	e.LogInfo("Anwendung gestartet", "pid", cmd.Process.Pid)
	return nil
}

// Validate überprüft eine App-Start-Aktion auf Gültigkeit
func (e *AppStartExecutor) Validate(action config.Action) error {
	// Path-Parameter überprüfen
	path, ok := action.Parameters["path"]
	if !ok {
		return fmt.Errorf("app_start-Aktion benötigt 'path' Parameter")
	}

	pathStr, ok := path.(string)
	if !ok {
		return fmt.Errorf("'path' Parameter muss ein String sein")
	}

	if pathStr == "" {
		return fmt.Errorf("'path' Parameter darf nicht leer sein")
	}

	// Plattformspezifische Pfad-Validierung
	if err := e.validatePath(pathStr); err != nil {
		return fmt.Errorf("ungültiger Pfad: %w", err)
	}

	// Args-Parameter überprüfen (falls vorhanden)
	if argsParam, ok := action.Parameters["args"]; ok {
		switch v := argsParam.(type) {
		case []interface{}:
			for i, arg := range v {
				if _, ok := arg.(string); !ok {
					return fmt.Errorf("arg %d muss ein String sein", i)
				}
			}
		case []string:
			// OK
		case string:
			// OK
		default:
			return fmt.Errorf("ungültiger 'args' Parameter: %v", argsParam)
		}
	}

	// Working-Dir-Parameter überprüfen (falls vorhanden)
	if dirParam, ok := action.Parameters["working_dir"]; ok {
		if dirStr, ok := dirParam.(string); ok {
			if dirStr != "" {
				if err := e.validatePath(dirStr); err != nil {
					return fmt.Errorf("ungültiges Arbeitsverzeichnis: %w", err)
				}
			}
		} else {
			return fmt.Errorf("'working_dir' Parameter muss ein String sein")
		}
	}

	return nil
}

// validatePath überprüft einen Pfad auf Gültigkeit
func (e *AppStartExecutor) validatePath(path string) error {
	// Plattformspezifische Validierung
	switch runtime.GOOS {
	case "windows":
		return e.validateWindowsPath(path)
	case "linux":
		return e.validateLinuxPath(path)
	default:
		return fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// validateWindowsPath überprüft einen Windows-Pfad
func (e *AppStartExecutor) validateWindowsPath(path string) error {
	// Grundlegende Windows-Pfad-Validierung
	if len(path) == 0 {
		return fmt.Errorf("pfad darf nicht leer sein")
	}

	// Verbotene Zeichen überprüfen
	forbiddenChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range forbiddenChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("pfad enthält verbotenes Zeichen: %s", char)
		}
	}

	// Relative Pfade sind erlaubt
	// Absolute Pfade sollten mit Laufwerk beginnen (z.B. C:\)
	if len(path) >= 2 && path[1] == ':' {
		// Absoluter Pfad
		if !strings.HasPrefix(path, "C:\\") && !strings.HasPrefix(path, "D:\\") {
			return fmt.Errorf("ungültiger Laufwerksbuchstabe in Pfad: %s", path)
		}
	}

	return nil
}

// validateLinuxPath überprüft einen Linux-Pfad
func (e *AppStartExecutor) validateLinuxPath(path string) error {
	// Grundlegende Linux-Pfad-Validierung
	if len(path) == 0 {
		return fmt.Errorf("pfad darf nicht leer sein")
	}

	// Verbotene Zeichen überprüfen
	forbiddenChars := []string{"\x00"}
	for _, char := range forbiddenChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("pfad enthält verbotenes Zeichen")
		}
	}

	// Absolute Pfade sollten mit / beginnen
	if strings.HasPrefix(path, "/") {
		// Absoluter Pfad - weitere Validierung möglich
		if strings.Contains(path, "..") {
			return fmt.Errorf("pfad enthält '..' was nicht erlaubt ist")
		}
	}

	return nil
}

// GetCommonApps gibt eine Liste häufiger Anwendungen zurück
func (e *AppStartExecutor) GetCommonApps() map[string]string {
	switch runtime.GOOS {
	case "windows":
		return map[string]string{
			"notepad":     "notepad.exe",
			"calculator":  "calc.exe",
			"explorer":    "explorer.exe",
			"cmd":         "cmd.exe",
			"powershell":  "powershell.exe",
			"chrome":      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"firefox":     "C:\\Program Files\\Mozilla Firefox\\firefox.exe",
			"obs":         "C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe",
			"discord":     "C:\\Users\\%USERNAME%\\AppData\\Local\\Discord\\app-1.0.9003\\Discord.exe",
			"spotify":     "C:\\Users\\%USERNAME%\\AppData\\Roaming\\Spotify\\Spotify.exe",
		}
	case "linux":
		return map[string]string{
			"gedit":       "gedit",
			"calculator":  "gnome-calculator",
			"terminal":    "gnome-terminal",
			"chrome":      "google-chrome",
			"firefox":     "firefox",
			"obs":         "obs",
			"discord":     "discord",
			"spotify":     "spotify",
			"vlc":         "vlc",
			"gimp":        "gimp",
		}
	default:
		return map[string]string{}
	}
}

// ResolveAppPath löst einen App-Namen in einen vollständigen Pfad auf
func (e *AppStartExecutor) ResolveAppPath(appName string) (string, error) {
	commonApps := e.GetCommonApps()
	
	if path, exists := commonApps[strings.ToLower(appName)]; exists {
		return path, nil
	}

	// Versuche den App-Namen direkt als Pfad zu verwenden
	return appName, nil
} 