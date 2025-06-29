// Package actions verwaltet die Ausführung von Systemaktionen basierend auf MIDI-Events.
// Es bietet plattformübergreifende Unterstützung für verschiedene Aktionstypen.
package actions

import (
	"fmt"
	"sync"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// Manager verwaltet die Ausführung von Systemaktionen
type Manager struct {
	config    *config.Config
	logger    utils.Logger
	executors map[string]Executor
	mutex     sync.RWMutex
}

// Executor definiert die Schnittstelle für Aktion-Ausführer
type Executor interface {
	Execute(action config.Action) error
	GetName() string
}

// NewManager erstellt einen neuen Action-Manager
func NewManager(cfg *config.Config, logger utils.Logger) (*Manager, error) {
	manager := &Manager{
		config:    cfg,
		logger:    logger,
		executors: make(map[string]Executor),
	}

	// Plattformspezifische Executors registrieren
	if err := manager.registerExecutors(); err != nil {
		return nil, fmt.Errorf("fehler beim Registrieren der Executors: %w", err)
	}

	return manager, nil
}

// registerExecutors registriert alle verfügbaren Aktion-Executors
func (m *Manager) registerExecutors() error {
	// Volume-Executor registrieren
	volumeExecutor, err := NewVolumeExecutor(m.logger)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Volume-Executors: %w", err)
	}
	m.registerExecutor(volumeExecutor)

	// App-Start-Executor registrieren
	appStartExecutor, err := NewAppStartExecutor(m.logger)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des App-Start-Executors: %w", err)
	}
	m.registerExecutor(appStartExecutor)

	// Tastenkombination-Executor registrieren
	keyCombinationExecutor, err := NewKeyCombinationExecutor(m.logger)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Key-Combination-Executors: %w", err)
	}
	m.registerExecutor(keyCombinationExecutor)

	// Audio-Quelle-Executor registrieren
	audioSourceExecutor, err := NewAudioSourceExecutor(m.logger)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Audio-Source-Executors: %w", err)
	}
	m.registerExecutor(audioSourceExecutor)

	m.logger.Info("Action-Executors registriert", "count", len(m.executors))
	return nil
}

// registerExecutor registriert einen einzelnen Executor
func (m *Manager) registerExecutor(executor Executor) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.executors[executor.GetName()] = executor
}

// Execute führt eine Aktion aus
func (m *Manager) Execute(action config.Action) error {
	m.mutex.RLock()
	executor, exists := m.executors[action.Type]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("kein Executor für Aktion-Typ '%s' gefunden", action.Type)
	}

	m.logger.Debug("Führe Aktion aus", "type", action.Type, "parameters", action.Parameters)

	if err := executor.Execute(action); err != nil {
		return fmt.Errorf("fehler beim Ausführen der Aktion '%s': %w", action.Type, err)
	}

	m.logger.Info("Aktion erfolgreich ausgeführt", "type", action.Type)
	return nil
}

// GetExecutor gibt einen Executor für einen bestimmten Typ zurück
func (m *Manager) GetExecutor(actionType string) (Executor, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	executor, exists := m.executors[actionType]
	return executor, exists
}

// GetAvailableTypes gibt alle verfügbaren Aktion-Typen zurück
func (m *Manager) GetAvailableTypes() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	types := make([]string, 0, len(m.executors))
	for actionType := range m.executors {
		types = append(types, actionType)
	}
	return types
}

// ValidateAction überprüft ob eine Aktion gültig ist
func (m *Manager) ValidateAction(action config.Action) error {
	executor, exists := m.GetExecutor(action.Type)
	if !exists {
		return fmt.Errorf("ungültiger Aktion-Typ: %s", action.Type)
	}

	// Spezifische Validierung durch den Executor
	if validator, ok := executor.(ActionValidator); ok {
		return validator.Validate(action)
	}

	return nil
}

// ActionValidator definiert die Schnittstelle für Aktion-Validierung
type ActionValidator interface {
	Validate(action config.Action) error
}

// BaseExecutor bietet grundlegende Funktionalität für Executors
type BaseExecutor struct {
	name   string
	logger utils.Logger
}

// NewBaseExecutor erstellt einen neuen Base-Executor
func NewBaseExecutor(name string, logger utils.Logger) BaseExecutor {
	return BaseExecutor{
		name:   name,
		logger: logger,
	}
}

// GetName gibt den Namen des Executors zurück
func (e *BaseExecutor) GetName() string {
	return e.name
}

// LogDebug loggt eine Debug-Nachricht
func (e *BaseExecutor) LogDebug(msg string, fields ...interface{}) {
	e.logger.Debug(msg, fields...)
}

// LogInfo loggt eine Info-Nachricht
func (e *BaseExecutor) LogInfo(msg string, fields ...interface{}) {
	e.logger.Info(msg, fields...)
}

// LogError loggt eine Fehler-Nachricht
func (e *BaseExecutor) LogError(msg string, fields ...interface{}) {
	e.logger.Error(msg, fields...)
} 