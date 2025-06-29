// Package utils bietet generische Hilfsfunktionen für MidiDaemon.
// Diese Datei enthält Plattform-Erkennung und Hilfsfunktionen.

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PlatformInfo enthält Informationen über die aktuelle Plattform
type PlatformInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
}

// GetPlatformInfo gibt Informationen über die aktuelle Plattform zurück
func GetPlatformInfo() PlatformInfo {
	hostname, _ := os.Hostname()

	return PlatformInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  runtime.Version(),
		Hostname: hostname,
	}
}

// IsWindows gibt zurück ob das System Windows ist
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux gibt zurück ob das System Linux ist
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsDarwin gibt zurück ob das System macOS ist
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsSupportedPlatform gibt zurück ob die aktuelle Plattform unterstützt wird
func IsSupportedPlatform() bool {
	return IsWindows() || IsLinux()
}

// GetExecutablePath gibt den Pfad zur ausführbaren Datei zurück
func GetExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("fehler beim Abrufen des Executable-Pfads: %w", err)
	}
	return filepath.Abs(exe)
}

// GetExecutableDir gibt das Verzeichnis der ausführbaren Datei zurück
func GetExecutableDir() (string, error) {
	exePath, err := GetExecutablePath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

// GetConfigDir gibt das Standard-Konfigurationsverzeichnis zurück
func GetConfigDir() (string, error) {
	if IsWindows() {
		// Windows: %APPDATA%\MidiDaemon
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA-Umgebungsvariable nicht gesetzt")
		}
		return filepath.Join(appData, "MidiDaemon"), nil
	} else if IsLinux() {
		// Linux: ~/.config/mididaemon
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("fehler beim Abrufen des Home-Verzeichnisses: %w", err)
		}
		return filepath.Join(homeDir, ".config", "mididaemon"), nil
	}

	return "", fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
}

// GetLogDir gibt das Standard-Log-Verzeichnis zurück
func GetLogDir() (string, error) {
	if IsWindows() {
		// Windows: %APPDATA%\MidiDaemon\logs
		configDir, err := GetConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(configDir, "logs"), nil
	} else if IsLinux() {
		// Linux: ~/.local/share/mididaemon/logs
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("fehler beim Abrufen des Home-Verzeichnisses: %w", err)
		}
		return filepath.Join(homeDir, ".local", "share", "mididaemon", "logs"), nil
	}

	return "", fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
}

// EnsureDir erstellt ein Verzeichnis falls es nicht existiert
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("fehler beim Erstellen des Verzeichnisses '%s': %w", path, err)
		}
	}
	return nil
}

// FileExists gibt zurück ob eine Datei existiert
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFile gibt zurück ob ein Pfad eine Datei ist
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir gibt zurück ob ein Pfad ein Verzeichnis ist
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize gibt die Größe einer Datei zurück
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("fehler beim Abrufen der Datei-Informationen: %w", err)
	}
	return info.Size(), nil
}

// NormalizePath normalisiert einen Pfad für die aktuelle Plattform
func NormalizePath(path string) string {
	if IsWindows() {
		// Windows-Pfade normalisieren
		path = strings.ReplaceAll(path, "/", "\\")
		path = strings.ReplaceAll(path, "\\\\", "\\")
	} else {
		// Unix-Pfade normalisieren
		path = strings.ReplaceAll(path, "\\", "/")
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}

// ExpandEnvVars erweitert Umgebungsvariablen in einem Pfad
func ExpandEnvVars(path string) string {
	if IsWindows() {
		// Windows-Umgebungsvariablen erweitern
		return os.ExpandEnv(path)
	} else {
		// Unix-Umgebungsvariablen erweitern
		return os.ExpandEnv(path)
	}
}

// GetEnvOrDefault gibt eine Umgebungsvariable zurück oder einen Standardwert
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt gibt eine Umgebungsvariable als Integer zurück oder einen Standardwert
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := parseInt(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetEnvAsBool gibt eine Umgebungsvariable als Boolean zurück oder einen Standardwert
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// GetSystemInfo gibt detaillierte Systeminformationen zurück
func GetSystemInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Plattform-Informationen
	platform := GetPlatformInfo()
	info["platform"] = platform

	// Go-Runtime-Informationen
	info["go_version"] = runtime.Version()
	info["go_os"] = runtime.GOOS
	info["go_arch"] = runtime.GOARCH
	info["go_compiler"] = runtime.Compiler
	info["num_cpu"] = runtime.NumCPU()
	info["num_goroutine"] = runtime.NumGoroutine()

	// Umgebungsvariablen
	info["env"] = map[string]string{
		"HOME":     os.Getenv("HOME"),
		"USER":     os.Getenv("USER"),
		"USERNAME": os.Getenv("USERNAME"),
		"PATH":     os.Getenv("PATH"),
	}

	// Verzeichnisse
	if configDir, err := GetConfigDir(); err == nil {
		info["config_dir"] = configDir
	}
	if logDir, err := GetLogDir(); err == nil {
		info["log_dir"] = logDir
	}
	if exeDir, err := GetExecutableDir(); err == nil {
		info["executable_dir"] = exeDir
	}

	return info
}

// FormatBytes formatiert Bytes in eine lesbare Größe
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseInt ist eine Hilfsfunktion zum Parsen von Strings zu Integers
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
