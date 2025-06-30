# MidiDaemon

MidiDaemon ist ein plattformübergreifendes Go-Programm, das MIDI-Controller-Eingaben in Systemaktionen umsetzt. Es unterstützt Windows und Linux und bietet eine intuitive Web-GUI zur Konfiguration.

## Features

### Core-Funktionalität
- **MIDI-Event-Verarbeitung**: Unterstützt Note On/Off, Control Change, Program Change
- **Systemaktionen**: Volume-Steuerung, App-Start, Tastenkombinationen, Audio-Quellen-Wechsel
- **Plattformübergreifend**: Windows und Linux Unterstützung
- **Konfigurierbar**: JSON-basierte Mapping-Konfiguration

### Web-GUI
- **Mapping-Verwaltung**: Intuitive Oberfläche für MIDI-Mappings
- **Controller-Erkennung**: Automatische Erkennung angeschlossener MIDI-Controller
- **Geräte-Informationen**: Hersteller, Modell, Typ und Fähigkeiten anzeigen
- **Intelligente Vorschläge**: Mapping-Vorschläge basierend auf Controller-Typ
- **Controller-Einstellungen**: Gerätespezifische Konfiguration

### Unterstützte Controller
- **Akai Professional**: MPK Mini, MPK249, MPX
- **Native Instruments**: Traktor Kontrol, Maschine
- **Behringer**: X32, XR18
- **Arturia**: KeyLab, BeatStep
- **Novation**: LaunchKey, LaunchPad
- **M-Audio**: Oxygen
- **Korg**: Nano-Serie
- **Roland**: A-Serie
- **Yamaha**: Motif
- **Generic Controller**: Automatische Erkennung unbekannter Geräte

## Installation

### Voraussetzungen
- Go 1.24 oder höher
- MIDI-Controller (optional)

### Build

```bash
# Repository klonen
git clone https://github.com/Xcruser/MidiDaemon.git
cd MidiDaemon

# Abhängigkeiten installieren
go mod tidy

# Für aktuelle Plattform
make build

# Für alle Plattformen
make build-all

# GUI bauen
make gui
```

### Plattformspezifische Builds

```bash
# Für Windows
make build-windows

# Für Linux
make build-linux

# Für macOS
make build-darwin

# GUI für alle Plattformen
make gui-all
```

## Verwendung

### Kommandozeile

```bash
# Mit Standard-Konfiguration
./mididaemon

# Mit spezifischer Konfigurationsdatei
./mididaemon -config myconfig.json

# Debug-Modus
./mididaemon -debug

# Beispiel-Konfiguration erzeugen
./mididaemon -generate-config
```

### Web-GUI

```bash
# GUI starten
./mididaemon-gui

# Oder mit Makefile
make run-gui
```

Die GUI ist dann unter `http://localhost:8080` erreichbar.

### Als systemd-Dienst

Eine Beispiel-Service-Datei befindet sich in `docs/mididaemon.service`. Nach dem
Kopieren nach `/etc/systemd/system/` kann der Dienst so aktiviert werden:

```bash
sudo systemctl enable --now mididaemon.service
```

#### GUI-Features

**Mappings-Tab:**
- Mapping-Übersicht und -Verwaltung
- Neues Mapping erstellen
- Bestehende Mappings bearbeiten/löschen
- Konfiguration speichern

**Controller-Tab:**
- Automatische Controller-Erkennung
- Geräte-Informationen anzeigen
- Controller-Einstellungen konfigurieren
- Intelligente Mapping-Vorschläge

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

## Controller-Erkennung

### Automatische Erkennung

Das System erkennt automatisch angeschlossene MIDI-Controller und zeigt:

- **Hersteller und Modell**: Identifikation des Geräts
- **Controller-Typ**: Keyboard, Pad, Knob, Slider, Mixer, DJ, etc.
- **Fähigkeiten**: Anzahl Tasten, Drehregler, Pads, etc.
- **Spezielle Features**: Display, Transport-Controls, etc.

### Controller-Einstellungen

Jeder Controller kann individuell konfiguriert werden:

- **Standard-Kanal**: MIDI-Kanal (0-15)
- **Velocity-sensitiv**: Velocity-Informationen verwenden
- **Druck-sensitiv**: Aftertouch/Expression verwenden
- **Automatisches Mapping**: Vorgeschlagene Mappings automatisch anwenden

### Intelligente Vorschläge

Das System generiert Mapping-Vorschläge basierend auf:

- **Controller-Typ**: Keyboard, Pad, Knob, etc.
- **Verfügbare Fähigkeiten**: Anzahl Tasten, Drehregler, etc.
- **Häufige Anwendungsfälle**: Volume-Control, Transport, etc.

## Verwendung

### Grundlegende Verwendung

```bash
# Mit Standard-Konfiguration
./mididaemon

# Mit spezifischer Konfigurationsdatei
./mididaemon -config myconfig.json

# Debug-Modus
./mididaemon -debug -verbose
```

### GUI-Verwendung

1. **GUI starten**: `./mididaemon-gui`
2. **Browser öffnen**: `http://localhost:8080`
3. **Controller erkennen**: Klicken Sie auf "Controller erkennen"
4. **Mappings erstellen**: Verwenden Sie die vorgeschlagenen Mappings oder erstellen Sie eigene
5. **Konfiguration speichern**: Klicken Sie auf "Speichern"

### Beispiele

#### Volume-Steuerung

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

#### Anwendung starten

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

#### Tastenkombination

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

### GUI-Entwicklung

```bash
# GUI im Development-Modus
make dev-gui

# GUI für alle Plattformen bauen
make gui-all
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

4. **Controller werden nicht erkannt**
   - Überprüfen Sie die MIDI-Verbindung
   - Stellen Sie sicher, dass der Controller eingeschaltet ist
   - Prüfen Sie die Treiber-Installation

### Logs

```bash
# Debug-Logs aktivieren
./mididaemon -debug -verbose

# Log-Level setzen
./mididaemon -log-level debug
```

## Lizenz

Dieses Projekt steht unter der MIT-Lizenz. Siehe [LICENSE](LICENSE) für Details.

## Beitragen

Beiträge sind willkommen! Bitte lesen Sie die [Contributing Guidelines](CONTRIBUTING.md) für Details.

## Changelog

### v1.0.0
- Initiale Version mit MIDI-Event-Verarbeitung
- Web-GUI für Mapping-Verwaltung
- Controller-Erkennung und -Verwaltung
- Intelligente Mapping-Vorschläge
- Plattformübergreifende Unterstützung
