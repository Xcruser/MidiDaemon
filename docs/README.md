# MidiDaemon – Entwickler- und Anwenderdokumentation

## Inhaltsverzeichnis

- [Einleitung](#einleitung)
- [Projektüberblick](#projektüberblick)
- [Installation & Build](#installation--build)
- [Konfiguration](#konfiguration)
- [MIDI-Mapping](#midi-mapping)
- [Aktionstypen](#aktionstypen)
- [Plattformdetails](#plattformdetails)
- [Entwicklung & Erweiterung](#entwicklung--erweiterung)
- [Testing & Debugging](#testing--debugging)
- [Docker & Deployment](#docker--deployment)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#roadmap)
- [Lizenz & Mitwirken](#lizenz--mitwirken)

---

## Einleitung

**MidiDaemon** ist ein plattformübergreifender Daemon, der MIDI-Controller-Eingaben in Systemaktionen umsetzt. Ziel ist es, beliebige MIDI-Controller als universelle Steuerzentrale für den PC zu nutzen – z. B. für Lautstärke, App-Start, Audioquellen, Tastenkombinationen u. v. m.

---

## Projektüberblick

- **Sprache:** Go (statisch gebaut, keine Laufzeit-Abhängigkeiten)
- **Plattformen:** Windows & Linux
- **Konfiguration:** JSON-basiert, beliebig erweiterbar
- **Aktionen:** Lautstärke, Programme starten, Audioquelle wechseln, Tastenkombinationen
- **Modular:** Neue Aktionen und Mappings einfach ergänzbar

### Projektstruktur

```
MidiDaemon/
├── cmd/mididaemon/main.go         # Einstiegspunkt
├── internal/
│   ├── config/                    # Konfigurationsverwaltung
│   ├── midi/                      # MIDI-Handler & Ports
│   └── actions/                   # Systemaktionen (plattformabhängig)
├── pkg/utils/                     # Logging, Plattformtools
├── config.json                    # Beispiel-Konfiguration
├── Makefile, Dockerfile, README.md
└── docs/                          # <--- Diese Dokumentation
```

---

## Installation & Build

### Voraussetzungen
- Go 1.21+
- Git
- (Optional) Docker

### Build (lokal)
```bash
git clone https://github.com/Xcruser/MidiDaemon.git
cd MidiDaemon
make build
./build/mididaemon -config config.json
```

### Cross-Compile
```bash
make build-windows
make build-linux
make build-all
```

### Docker
```bash
docker build -t mididaemon .
docker run --rm -it mididaemon
```

---

## Konfiguration

Die Datei `config.json` steuert das Verhalten. Sie enthält:
- MIDI-Port, Kanal, Timeout
- Logging, Action-Delay
- Eine Liste von Mappings (MIDI-Event → Aktion)

### Beispiel
```json
{
  "midi": { "input_port": "", "channel": -1, "timeout": 30 },
  "general": { "log_level": "info", "auto_restart": true, "action_delay": 100 },
  "mappings": [
    {
      "name": "Volume Up",
      "event": { "type": "control_change", "controller": 7, "value": 64 },
      "action": { "type": "volume", "parameters": { "direction": "up", "percent": 5 } }
    }
  ]
}
```

---

## MIDI-Mapping

Jedes Mapping besteht aus:
- **event**: MIDI-Event (z. B. Note, Controller, Program Change)
- **action**: Systemaktion (z. B. Volume, App-Start)
- **enabled**: Aktiviert/Deaktiviert

### Event-Typen
- `note_on`, `note_off`, `control_change`, `program_change`

### Beispiel-Mapping
```json
{
  "name": "Start OBS",
  "event": { "type": "note_on", "note": 60, "velocity": 100 },
  "action": { "type": "app_start", "parameters": { "path": "obs", "args": ["--startstreaming"] } }
}
```

---

## Aktionstypen

### Volume
```json
{
  "type": "volume",
  "parameters": { "direction": "up|down|set|mute|unmute", "percent": 5, "volume": 50 }
}
```

### App-Start
```json
{
  "type": "app_start",
  "parameters": { "path": "notepad.exe", "args": ["foo"] }
}
```

### Tastenkombination
```json
{
  "type": "key_combination",
  "parameters": { "keys": ["CTRL", "C"], "type": "combination|sequence|hold|text" }
}
```

### Audio-Quelle
```json
{
  "type": "audio_source",
  "parameters": { "source": "speakers", "type": "switch|mute|unmute|volume|cycle", "volume": 50 }
}
```

---

## Plattformdetails

### Windows
- MIDI: Windows MIDI API (Platzhalter für gomidi)
- Volume: Core Audio API
- Tastatur: Windows API

### Linux
- MIDI: ALSA (Platzhalter für gomidi)
- Volume: ALSA/PulseAudio
- Tastatur: X11/uinput

---

## Entwicklung & Erweiterung

- **Neue Aktion:** In `internal/actions/` neuen Executor anlegen und in `manager.go` registrieren.
- **Neues Mapping:** Einfach in `config.json` ergänzen.
- **Tests:** Siehe Makefile (`make test`)
- **Logging:** Über `pkg/utils/logger.go` steuerbar.

---

## Testing & Debugging

- **Debug-Log:** `./mididaemon -verbose`
- **Tests:** `make test`
- **Coverage:** `make test-coverage`
- **Logs:** Standardausgabe oder Datei (umleiten mit `> log.txt`)

---

## Docker & Deployment

- **Build:** `docker build -t mididaemon .`
- **Run:**  `docker run --rm -it mididaemon`
- **Konfig:** `/root/config.json` im Container

---

## Troubleshooting

- **MIDI-Port nicht gefunden:** Prüfe Verkabelung, Portname, Berechtigungen
- **Aktion wird nicht ausgeführt:** Prüfe Mapping, Log-Ausgabe, Konfiguration
- **Linux Berechtigungen:** User ggf. zur `audio`-Gruppe hinzufügen

---

## Roadmap

- macOS-Support
- Web-UI für Konfiguration
- Hot-Reload der Config
- Plugin-System
- MIDI-Learn-Modus
- Systemd/Windows-Service

---

## Lizenz & Mitwirken

- **Lizenz:** MIT
- **Mitwirken:** Pull Requests & Issues willkommen!
- **Autor:** [Xcruser](https://github.com/Xcruser)

---

*Letzte Aktualisierung: $(date +%Y-%m-%d)* 