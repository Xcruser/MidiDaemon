// Package midi verwaltet MIDI-Eingaben und leitet sie an die entsprechenden Aktionen weiter.
// Es unterstützt verschiedene MIDI-Event-Typen und bietet eine plattformübergreifende Schnittstelle.
package midi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Xcruser/MidiDaemon/internal/actions"
	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

// Handler verwaltet MIDI-Eingaben und leitet sie an Aktionen weiter
type Handler struct {
	config     *config.Config
	logger     utils.Logger
	actionMgr  *actions.Manager
	port       MIDIPort
	eventChan  chan MIDIEvent
	done       chan struct{}
	mutex      sync.RWMutex
	isRunning  bool
}

// MIDIEvent repräsentiert ein empfangenes MIDI-Event
type MIDIEvent struct {
	Type      string // "note_on", "note_off", "control_change", "program_change"
	Channel   int    // MIDI-Kanal (0-15)
	Note      int    // MIDI-Note (0-127)
	Controller int   // Controller-Nummer (0-127)
	Program   int    // Program-Nummer (0-127)
	Velocity  int    // Velocity (0-127)
	Value     int    // Controller-Wert (0-127)
	Timestamp time.Time
}

// MIDIPort definiert die Schnittstelle für MIDI-Ports
type MIDIPort interface {
	Open(portName string) error
	Close() error
	ReadEvents() (<-chan MIDIEvent, error)
	GetPortNames() ([]string, error)
}

// NewHandler erstellt einen neuen MIDI-Handler
func NewHandler(cfg *config.Config, logger utils.Logger) (*Handler, error) {
	// Action-Manager erstellen
	actionMgr, err := actions.NewManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des Action-Managers: %w", err)
	}

	// Plattformspezifischen MIDI-Port erstellen
	port, err := newMIDIPort()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Erstellen des MIDI-Ports: %w", err)
	}

	handler := &Handler{
		config:    cfg,
		logger:    logger,
		actionMgr: actionMgr,
		port:      port,
		eventChan: make(chan MIDIEvent, 100),
		done:      make(chan struct{}),
	}

	return handler, nil
}

// Start startet den MIDI-Handler
func (h *Handler) Start(ctx context.Context) error {
	h.mutex.Lock()
	if h.isRunning {
		h.mutex.Unlock()
		return fmt.Errorf("handler läuft bereits")
	}
	h.isRunning = true
	h.mutex.Unlock()

	h.logger.Info("MIDI-Handler wird gestartet")

	// MIDI-Port öffnen
	portName := h.config.MIDI.InputPort
	if portName == "" {
		// Ersten verfügbaren Port verwenden
		ports, err := h.port.GetPortNames()
		if err != nil {
			return fmt.Errorf("fehler beim Abrufen der MIDI-Ports: %w", err)
		}
		if len(ports) == 0 {
			return fmt.Errorf("keine MIDI-Ports verfügbar")
		}
		portName = ports[0]
		h.logger.Info("Verwende ersten verfügbaren MIDI-Port", "port", portName)
	}

	if err := h.port.Open(portName); err != nil {
		return fmt.Errorf("fehler beim Öffnen des MIDI-Ports '%s': %w", portName, err)
	}

	h.logger.Info("MIDI-Port geöffnet", "port", portName)

	// Event-Stream starten
	eventStream, err := h.port.ReadEvents()
	if err != nil {
		return fmt.Errorf("fehler beim Starten des Event-Streams: %w", err)
	}

	// Event-Verarbeitung in separater Goroutine
	go h.processEvents(ctx, eventStream)

	// Auf Context-Cancellation warten
	<-ctx.Done()
	h.logger.Info("MIDI-Handler wird beendet")

	return nil
}

// Close beendet den MIDI-Handler
func (h *Handler) Close() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		return nil
	}

	h.logger.Info("MIDI-Handler wird geschlossen")

	// Port schließen
	if h.port != nil {
		if err := h.port.Close(); err != nil {
			h.logger.Error("Fehler beim Schließen des MIDI-Ports", "error", err)
		}
	}

	// Channels schließen
	close(h.done)
	close(h.eventChan)

	h.isRunning = false
	return nil
}

// processEvents verarbeitet eingehende MIDI-Events
func (h *Handler) processEvents(ctx context.Context, eventStream <-chan MIDIEvent) {
	for {
		select {
		case event, ok := <-eventStream:
			if !ok {
				h.logger.Info("MIDI-Event-Stream wurde geschlossen")
				return
			}
			h.handleEvent(event)

		case <-ctx.Done():
			h.logger.Info("Event-Verarbeitung wird beendet")
			return

		case <-h.done:
			h.logger.Info("Event-Verarbeitung wird beendet")
			return
		}
	}
}

// handleEvent verarbeitet ein einzelnes MIDI-Event
func (h *Handler) handleEvent(event MIDIEvent) {
	// Kanal-Filterung
	if h.config.MIDI.Channel != -1 && event.Channel != h.config.MIDI.Channel {
		return
	}

	h.logger.Debug("MIDI-Event empfangen",
		"type", event.Type,
		"channel", event.Channel,
		"note", event.Note,
		"controller", event.Controller,
		"velocity", event.Velocity,
		"value", event.Value,
	)

	// Passende Mappings finden und ausführen
	for _, mapping := range h.config.Mappings {
		if !mapping.Enabled {
			continue
		}

		if h.matchesMapping(event, mapping.Event) {
			h.logger.Info("Mapping gefunden", "name", mapping.Name)
			
			// Aktion in separater Goroutine ausführen
			go func(m config.Mapping) {
				if err := h.actionMgr.Execute(m.Action); err != nil {
					h.logger.Error("Fehler beim Ausführen der Aktion",
						"mapping", m.Name,
						"action", m.Action.Type,
						"error", err,
					)
				}
			}(mapping)

			// Verzögerung zwischen Aktionen
			if h.config.General.ActionDelay > 0 {
				time.Sleep(time.Duration(h.config.General.ActionDelay) * time.Millisecond)
			}
		}
	}
}

// matchesMapping überprüft ob ein MIDI-Event zu einem Mapping passt
func (h *Handler) matchesMapping(event MIDIEvent, mappingEvent config.MIDIEvent) bool {
	// Event-Typ überprüfen
	if event.Type != mappingEvent.Type {
		return false
	}

	switch event.Type {
	case "note_on", "note_off":
		// Note überprüfen
		if event.Note != mappingEvent.Note {
			return false
		}
		// Velocity-Schwellwert überprüfen (falls definiert)
		if mappingEvent.Velocity > 0 && event.Velocity < mappingEvent.Velocity {
			return false
		}

	case "control_change":
		// Controller überprüfen
		if event.Controller != mappingEvent.Controller {
			return false
		}
		// Wert-Schwellwert überprüfen (falls definiert)
		if mappingEvent.Value > 0 && event.Value < mappingEvent.Value {
			return false
		}

	case "program_change":
		// Program überprüfen
		if event.Program != mappingEvent.Program {
			return false
		}
	}

	return true
}

// GetPortNames gibt eine Liste verfügbarer MIDI-Ports zurück
func (h *Handler) GetPortNames() ([]string, error) {
	return h.port.GetPortNames()
}

// IsRunning gibt zurück ob der Handler läuft
func (h *Handler) IsRunning() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.isRunning
} 