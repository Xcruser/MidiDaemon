# MidiDaemon

Ein plattformübergreifender MIDI-Controller-Daemon, der MIDI-Eingaben in Systemaktionen umsetzt.

## Übersicht

MidiDaemon ist ein in Go geschriebener Daemon, der MIDI-Controller-Eingaben empfängt und diese in konfigurierbare Systemaktionen umsetzt. Unterstützt werden Windows und Linux.

### Features

- **Plattformübergreifend**: Unterstützt Windows und Linux
- **Konfigurierbar**: JSON-basierte Mapping-Konfiguration
- **Grafische Benutzeroberfläche**: Benutzerfreundliche GUI zur Konfiguration
- **Verschiedene Aktionstypen**:
  - Lautstärkesteuerung (Volume Up/Down, Mute/Unmute)
  - Anwendungen starten
  - Tastenkombinationen senden
  - Audioquellen wechseln
- **Keine Laufzeit-Abhängigkeiten**: Statisch gebautes Binary
- **Robust**: Graceful Shutdown und Fehlerbehandlung
- **Logging**: Umfassendes Logging-System

### Grafische Benutzeroberfläche

MidiDaemon bietet eine moderne, plattformübergreifende GUI zur einfachen Konfiguration von MIDI-Mappings:

- **Mapping-Übersicht**: Anzeige aller konfigurierten Mappings
- **Mapping-Editor**: Einfaches Hinzufügen, Bearbeiten und Löschen von Mappings
- **Live-Validierung**: Sofortige Überprüfung der Eingaben
- **MIDI Learn Modus**: Automatische Erkennung von MIDI-Eingaben (geplant)
- **Konfigurationsspeicherung**: Direktes Speichern in die config.json

```bash
# GUI starten
make run-gui

# GUI bauen
make build-gui
```

Weitere Informationen zur GUI finden Sie in der [GUI-Dokumentation](docs/GUI.md).

## Installation

### Voraussetzungen

- Go 1.21 oder höher
- MIDI-Controller oder MIDI-Interface
- Windows oder Linux

### Build

```bash
# Repository klonen
git clone https://github.com/Xcruser/MidiDaemon.git
cd MidiDaemon

# Abhängigkeiten installieren
go mod download

# Binary bauen
make build

# Oder direkt mit Go
go build -o mididaemon cmd/mididaemon/main.go
```

### Cross-Compilation

```bash
# Für Windows
make build-windows

# Für Linux
make build-linux

# Für macOS
make build-darwin

# Alle Plattformen
make build-all
```

## Konfiguration

MidiDaemon verwendet eine JSON-Konfigurationsdatei (`config.json`) für die Mapping-Definitionen.

### Beispiel-Konfiguration

```json
{
  "midi": {
    "input_port": "",
    "channel": -1,
    "timeout": 30
  },
  "general": {
    "log_level": "info",
    "auto_restart": true,
    "action_delay": 100
  },
  "mappings": [
    {
      "name": "Volume Up",
      "enabled": true,
      "event": {
        "type": "control_change",
        "controller": 7,
        "value": 64
      },
      "action": {
        "type": "volume",
        "parameters": {
          "direction": "up",
          "percent": 5
        }
      }
    }
  ]
}
```

### MIDI-Events

Unterstützte MIDI-Event-Typen:

- **note_on**: MIDI-Note gedrückt
- **note_off**: MIDI-Note losgelassen
- **control_change**: MIDI-Controller geändert
- **program_change**: MIDI-Program gewechselt

### Aktionstypen

#### Volume-Aktionen

```json
{
  "type": "volume",
  "parameters": {
    "direction": "up|down|set|mute|unmute",
    "percent": 5,
    "volume": 50
  }
}
```

#### App-Start-Aktionen

```json
{
  "type": "app_start",
  "parameters": {
    "path": "notepad.exe",
    "args": ["--startstreaming"],
    "working_dir": "C:\\Programs"
  }
}
```

#### Tastenkombination-Aktionen

```json
{
  "type": "key_combination",
  "parameters": {
    "keys": ["CTRL", "C"],
    "type": "combination|sequence|hold|text",
    "delay": 50,
    "duration": 1000
  }
}
```

#### Audio-Quelle-Aktionen

```json
{
  "type": "audio_source",
  "parameters": {
    "source": "speakers",
    "type": "switch|mute|unmute|volume|cycle",
    "volume": 50
  }
}
```

## Verwendung

### Grundlegende Verwendung

```bash
# Mit Standard-Konfiguration
./mididaemon

# Mit benutzerdefinierter Konfiguration
./mididaemon -config myconfig.json

# Mit Debug-Ausgabe
./mididaemon -verbose

# Version anzeigen
./mididaemon -version
```

### Kommandozeilen-Optionen

- `-config string`: Pfad zur Konfigurationsdatei (default: "config.json")
- `-verbose`: Ausführliche Logging-Ausgabe
- `-version`: Version anzeigen

### Entwicklung

```bash
# Entwicklungs-Setup
make dev-setup

# Code formatieren
make fmt

# Tests ausführen
make test

# Linting
make lint

# Vollständiger Check
make check
```

## Projektstruktur

```
MidiDaemon/
├── cmd/
│   └── mididaemon/
│       └── main.go              # Einstiegspunkt
├── internal/
│   ├── config/
│   │   └── config.go            # Konfigurationsverwaltung
│   ├── midi/
│   │   ├── handler.go           # MIDI-Event-Handler
│   │   └── port.go              # Plattformspezifische MIDI-Ports
│   └── actions/
│       ├── manager.go           # Action-Manager
│       ├── volume.go            # Volume-Executor
│       ├── app_start.go         # App-Start-Executor
│       ├── key_combination.go   # Key-Combination-Executor
│       └── audio_source.go      # Audio-Source-Executor
├── pkg/
│   └── utils/
│       ├── logger.go            # Logging-System
│       └── platform.go          # Plattform-Erkennung
├── config.json                  # Beispiel-Konfiguration
├── Makefile                     # Build-System
├── go.mod                       # Go-Module
└── README.md                    # Diese Datei
```

## MIDI-Mapping-Beispiele

### Grundlegende Lautstärkesteuerung

```json
{
  "name": "Volume Control",
  "event": {
    "type": "control_change",
    "controller": 7
  },
  "action": {
    "type": "volume",
    "parameters": {
      "direction": "set",
      "volume": "{{value}}"
    }
  }
}
```

### Anwendung starten

```json
{
  "name": "Start OBS",
  "event": {
    "type": "note_on",
    "note": 60
  },
  "action": {
    "type": "app_start",
    "parameters": {
      "path": "obs",
      "args": ["--startstreaming"]
    }
  }
}
```

### Tastenkombination

```json
{
  "name": "Screenshot",
  "event": {
    "type": "note_on",
    "note": 62
  },
  "action": {
    "type": "key_combination",
    "parameters": {
      "keys": ["CTRL", "SHIFT", "S"]
    }
  }
}
```

## Plattform-spezifische Hinweise

### Windows

- Verwendet Windows MIDI API
- Unterstützt alle gängigen MIDI-Controller
- Volume-Steuerung über Windows Core Audio API
- Tastatureingaben über Windows API

### Linux

- Verwendet ALSA MIDI
- Volume-Steuerung über ALSA oder PulseAudio
- Tastatureingaben über X11 oder uinput
- Benötigt möglicherweise zusätzliche Berechtigungen

## Entwicklung

### Abhängigkeiten hinzufügen

```bash
go get github.com/example/package
go mod tidy
```

### Tests schreiben

```bash
# Tests ausführen
go test ./...

# Tests mit Coverage
make test-coverage
```

### Debugging

```bash
# Debug-Version bauen
make build-debug

# Mit Debug-Ausgabe ausführen
make run-debug
```

## Troubleshooting

### Häufige Probleme

1. **MIDI-Port nicht gefunden**
   - Überprüfen Sie die MIDI-Verbindung
   - Verwenden Sie `-verbose` für Debug-Ausgabe

2. **Berechtigungsfehler (Linux)**
   - Fügen Sie Benutzer zur `audio`-Gruppe hinzu
   - Überprüfen Sie ALSA-Berechtigungen

3. **Aktionen werden nicht ausgeführt**
   - Überprüfen Sie die Konfigurationsdatei
   - Prüfen Sie die Logs auf Fehler

### Logs

MidiDaemon erstellt detaillierte Logs. Bei Problemen:

```bash
# Mit Debug-Ausgabe starten
./mididaemon -verbose

# Logs in Datei schreiben
./mididaemon -config config.json > mididaemon.log 2>&1
```

## Lizenz

Dieses Projekt ist unter der MIT-Lizenz lizenziert. Siehe [LICENSE](LICENSE) für Details.

## Beitragen

Beiträge sind willkommen! Bitte:

1. Fork das Repository
2. Erstellen Sie einen Feature-Branch
3. Committen Sie Ihre Änderungen
4. Pushen Sie den Branch
5. Erstellen Sie einen Pull Request

### Entwicklungsrichtlinien

- Folgen Sie den Go-Coding-Standards
- Schreiben Sie Tests für neue Features
- Aktualisieren Sie die Dokumentation
- Verwenden Sie `make check` vor dem Commit

## Roadmap

- [ ] macOS-Unterstützung
- [ ] Web-Interface für Konfiguration
- [ ] Plugin-System für benutzerdefinierte Aktionen
- [ ] MIDI-Learning-Modus
- [ ] Hot-Reload der Konfiguration
- [ ] Systemd-Service-Integration (Linux)
- [ ] Windows-Service-Integration

## Support

Bei Fragen oder Problemen:

1. Überprüfen Sie die [Issues](https://github.com/Xcruser/MidiDaemon/issues)
2. Erstellen Sie ein neues Issue mit detaillierten Informationen
3. Fügen Sie Logs und Konfiguration bei

## Credits

Entwickelt von [Xcruser](https://github.com/Xcruser)

---

**Hinweis**: Dieses Projekt ist in der Entwicklung. Die API kann sich noch ändern.
