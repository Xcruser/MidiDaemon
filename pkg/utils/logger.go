// Package utils bietet generische Hilfsfunktionen für MidiDaemon.
// Diese Datei enthält das Logging-System.

package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// Logger definiert die Schnittstelle für das Logging-System
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// LogLevel definiert die verschiedenen Log-Level
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String gibt den String-Repräsentation eines LogLevel zurück
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parst einen String zu einem LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo // Standard
	}
}

// ConsoleLogger implementiert ein einfaches Console-Logging
type ConsoleLogger struct {
	level     LogLevel
	verbose   bool
	timestamp bool
}

// NewLogger erstellt einen neuen Logger
func NewLogger(verbose bool) Logger {
	level := LevelInfo
	if verbose {
		level = LevelDebug
	}

	return &ConsoleLogger{
		level:     level,
		verbose:   verbose,
		timestamp: true,
	}
}

// NewLoggerWithLevel erstellt einen neuen Logger mit spezifischem Level
func NewLoggerWithLevel(level LogLevel, verbose bool) Logger {
	return &ConsoleLogger{
		level:     level,
		verbose:   verbose,
		timestamp: true,
	}
}

// Debug loggt eine Debug-Nachricht
func (l *ConsoleLogger) Debug(msg string, fields ...interface{}) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info loggt eine Info-Nachricht
func (l *ConsoleLogger) Info(msg string, fields ...interface{}) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn loggt eine Warn-Nachricht
func (l *ConsoleLogger) Warn(msg string, fields ...interface{}) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error loggt eine Error-Nachricht
func (l *ConsoleLogger) Error(msg string, fields ...interface{}) {
	if l.level <= LevelError {
		l.log(LevelError, msg, fields...)
	}
}

// Fatal loggt eine Fatal-Nachricht und beendet das Programm
func (l *ConsoleLogger) Fatal(msg string, fields ...interface{}) {
	if l.level <= LevelFatal {
		l.log(LevelFatal, msg, fields...)
		os.Exit(1)
	}
}

// log ist die interne Logging-Funktion
func (l *ConsoleLogger) log(level LogLevel, msg string, fields ...interface{}) {
	// Timestamp erstellen
	timestamp := ""
	if l.timestamp {
		timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	// Level-String
	levelStr := level.String()

	// Caller-Information (nur bei Debug-Level)
	caller := ""
	if l.verbose && level == LevelDebug {
		if pc, file, line, ok := runtime.Caller(2); ok {
			funcName := runtime.FuncForPC(pc).Name()
			// Nur den Funktionsnamen extrahieren
			parts := strings.Split(funcName, ".")
			if len(parts) > 0 {
				funcName = parts[len(parts)-1]
			}
			caller = fmt.Sprintf(" [%s:%d:%s]", getFileName(file), line, funcName)
		}
	}

	// Basis-Log-Nachricht
	logMsg := fmt.Sprintf("[%s] %s%s: %s", levelStr, timestamp, caller, msg)

	// Felder hinzufügen
	if len(fields) > 0 {
		fieldStr := l.formatFields(fields...)
		if fieldStr != "" {
			logMsg += " " + fieldStr
		}
	}

	// Ausgabe
	switch level {
	case LevelDebug:
		log.Printf("\033[36m%s\033[0m", logMsg) // Cyan
	case LevelInfo:
		log.Printf("\033[32m%s\033[0m", logMsg) // Green
	case LevelWarn:
		log.Printf("\033[33m%s\033[0m", logMsg) // Yellow
	case LevelError:
		log.Printf("\033[31m%s\033[0m", logMsg) // Red
	case LevelFatal:
		log.Printf("\033[35m%s\033[0m", logMsg) // Magenta
	default:
		log.Print(logMsg)
	}
}

// formatFields formatiert die Felder für die Ausgabe
func (l *ConsoleLogger) formatFields(fields ...interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		} else {
			// Ungerade Anzahl von Feldern - letztes Feld ohne Wert
			parts = append(parts, fmt.Sprintf("%v", fields[i]))
		}
	}

	return strings.Join(parts, " ")
}

// getFileName gibt nur den Dateinamen zurück (ohne Pfad)
func getFileName(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// FileLogger implementiert ein File-Logging
type FileLogger struct {
	*ConsoleLogger
	file *os.File
}

// NewFileLogger erstellt einen neuen File-Logger
func NewFileLogger(filename string, level LogLevel, verbose bool) (Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Öffnen der Log-Datei: %w", err)
	}

	return &FileLogger{
		ConsoleLogger: &ConsoleLogger{
			level:     level,
			verbose:   verbose,
			timestamp: true,
		},
		file: file,
	}, nil
}

// Close schließt die Log-Datei
func (l *FileLogger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// MultiLogger implementiert ein Multi-Destination-Logging
type MultiLogger struct {
	loggers []Logger
}

// NewMultiLogger erstellt einen neuen Multi-Logger
func NewMultiLogger(loggers ...Logger) Logger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// Debug loggt eine Debug-Nachricht an alle Logger
func (l *MultiLogger) Debug(msg string, fields ...interface{}) {
	for _, logger := range l.loggers {
		logger.Debug(msg, fields...)
	}
}

// Info loggt eine Info-Nachricht an alle Logger
func (l *MultiLogger) Info(msg string, fields ...interface{}) {
	for _, logger := range l.loggers {
		logger.Info(msg, fields...)
	}
}

// Warn loggt eine Warn-Nachricht an alle Logger
func (l *MultiLogger) Warn(msg string, fields ...interface{}) {
	for _, logger := range l.loggers {
		logger.Warn(msg, fields...)
	}
}

// Error loggt eine Error-Nachricht an alle Logger
func (l *MultiLogger) Error(msg string, fields ...interface{}) {
	for _, logger := range l.loggers {
		logger.Error(msg, fields...)
	}
}

// Fatal loggt eine Fatal-Nachricht an alle Logger und beendet das Programm
func (l *MultiLogger) Fatal(msg string, fields ...interface{}) {
	for _, logger := range l.loggers {
		logger.Fatal(msg, fields...)
	}
}

// NullLogger implementiert ein Null-Logging (für Tests)
type NullLogger struct{}

// NewNullLogger erstellt einen neuen Null-Logger
func NewNullLogger() Logger {
	return &NullLogger{}
}

// Debug tut nichts
func (l *NullLogger) Debug(msg string, fields ...interface{}) {}

// Info tut nichts
func (l *NullLogger) Info(msg string, fields ...interface{}) {}

// Warn tut nichts
func (l *NullLogger) Warn(msg string, fields ...interface{}) {}

// Error tut nichts
func (l *NullLogger) Error(msg string, fields ...interface{}) {}

// Fatal tut nichts (beendet nicht das Programm)
func (l *NullLogger) Fatal(msg string, fields ...interface{}) {}
