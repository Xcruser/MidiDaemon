// Package midi verwaltet MIDI-Eingaben und leitet sie an die entsprechenden Aktionen weiter.
// Diese Datei enthält plattformspezifische Implementierungen für MIDI-Ports.

package midi

import (
	"fmt"
	"runtime"
	"time"
)

// NewMIDIPort erstellt einen plattformspezifischen MIDI-Port
func NewMIDIPort() (MIDIPort, error) {
	switch runtime.GOOS {
	case "windows":
		return newWindowsMIDIPort()
	case "linux":
		return newLinuxMIDIPort()
	default:
		return nil, fmt.Errorf("plattform %s wird nicht unterstützt", runtime.GOOS)
	}
}

// newMIDIPort erstellt einen plattformspezifischen MIDI-Port (interne Funktion)
func newMIDIPort() (MIDIPort, error) {
	return NewMIDIPort()
}

// Windows-spezifische Implementierung
type windowsMIDIPort struct {
	portName string
	isOpen   bool
}

func newWindowsMIDIPort() (MIDIPort, error) {
	return &windowsMIDIPort{}, nil
}

func (p *windowsMIDIPort) Open(portName string) error {
	p.portName = portName
	p.isOpen = true
	// TODO: Windows MIDI API Implementierung
	return nil
}

func (p *windowsMIDIPort) Close() error {
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	// TODO: Windows MIDI API Implementierung
	return nil
}

func (p *windowsMIDIPort) ReadEvents() (<-chan MIDIEvent, error) {
	if !p.isOpen {
		return nil, fmt.Errorf("port ist nicht geöffnet")
	}

	eventChan := make(chan MIDIEvent, 100)

	// Platzhalter-Implementierung für Tests
	go func() {
		time.Sleep(5 * time.Second)
		select {
		case eventChan <- MIDIEvent{
			Type:      "note_on",
			Channel:   0,
			Note:      60,
			Velocity:  100,
			Timestamp: time.Now(),
		}:
		default:
		}
	}()

	return eventChan, nil
}

func (p *windowsMIDIPort) GetPortNames() ([]string, error) {
	return []string{
		"MIDI-Controller",
		"USB-MIDI-Interface",
		"Virtual-MIDI-Port",
		"Microsoft GS Wavetable Synth",
	}, nil
}

// Linux-spezifische Implementierung
type linuxMIDIPort struct {
	portName string
	isOpen   bool
}

func newLinuxMIDIPort() (MIDIPort, error) {
	return &linuxMIDIPort{}, nil
}

func (p *linuxMIDIPort) Open(portName string) error {
	p.portName = portName
	p.isOpen = true
	// TODO: ALSA MIDI API Implementierung
	return nil
}

func (p *linuxMIDIPort) Close() error {
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	// TODO: ALSA MIDI API Implementierung
	return nil
}

func (p *linuxMIDIPort) ReadEvents() (<-chan MIDIEvent, error) {
	if !p.isOpen {
		return nil, fmt.Errorf("port ist nicht geöffnet")
	}

	eventChan := make(chan MIDIEvent, 100)

	// Platzhalter-Implementierung für Tests
	go func() {
		time.Sleep(5 * time.Second)
		select {
		case eventChan <- MIDIEvent{
			Type:      "note_on",
			Channel:   0,
			Note:      60,
			Velocity:  100,
			Timestamp: time.Now(),
		}:
		default:
		}
	}()

	return eventChan, nil
}

func (p *linuxMIDIPort) GetPortNames() ([]string, error) {
	return []string{
		"ALSA MIDI Port 0",
		"USB MIDI Interface",
		"Virtual MIDI Port",
		"FLUID Synth",
		"Timidity",
	}, nil
}

// Mock-Implementierung für Tests und Entwicklung
type mockMIDIPort struct {
	portName  string
	isOpen    bool
	eventChan chan MIDIEvent
}

func newMockMIDIPort() (MIDIPort, error) {
	return &mockMIDIPort{
		eventChan: make(chan MIDIEvent, 100),
	}, nil
}

func (p *mockMIDIPort) Open(portName string) error {
	p.portName = portName
	p.isOpen = true
	return nil
}

func (p *mockMIDIPort) Close() error {
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	close(p.eventChan)
	return nil
}

func (p *mockMIDIPort) ReadEvents() (<-chan MIDIEvent, error) {
	if !p.isOpen {
		return nil, fmt.Errorf("port ist nicht geöffnet")
	}
	return p.eventChan, nil
}

func (p *mockMIDIPort) GetPortNames() ([]string, error) {
	return []string{"Mock MIDI Port"}, nil
}

// SendMockEvent sendet ein Test-Event (nur für Mock-Implementierung)
func (p *mockMIDIPort) SendMockEvent(event MIDIEvent) {
	if p.isOpen {
		p.eventChan <- event
	}
}
