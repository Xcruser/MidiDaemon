# MidiDaemon Web-GUI

Die MidiDaemon Web-GUI bietet eine benutzerfreundliche Oberfläche zur Konfiguration von MIDI-Controller-Mappings. Sie läuft als Web-Server und kann über jeden modernen Browser aufgerufen werden.

## Features

### Mapping-Verwaltung
- **Mapping-Übersicht**: Alle konfigurierten Mappings anzeigen
- **Mapping-Editor**: Neue Mappings erstellen und bestehende bearbeiten
- **Live-Validierung**: Sofortige Überprüfung der Eingaben
- **Speichern/Laden**: Konfiguration in `config.json` speichern

### Controller-Erkennung
- **Automatische Erkennung**: MIDI-Controller automatisch erkennen
- **Geräte-Informationen**: Hersteller, Modell, Typ und Fähigkeiten anzeigen
- **Controller-Einstellungen**: Gerätespezifische Konfiguration
- **Vorgeschlagene Mappings**: Intelligente Mapping-Vorschläge basierend auf Controller-Typ

## Installation und Start

### Voraussetzungen
- Go 1.24 oder höher
- MIDI-Controller (optional für Tests)

### Kompilieren
```bash
# Im Projektverzeichnis
cd cmd/mididaemon-gui
go build -o mididaemon-gui.exe .
```

### Starten
```bash
# GUI starten
./mididaemon-gui.exe

# Oder mit Makefile
make gui
```

Die GUI ist dann unter `http://localhost:8080` erreichbar.

## Verwendung

### Mappings-Tab

#### Mapping erstellen
1. Klicken Sie auf "Neues Mapping"
2. Füllen Sie die Felder aus:
   - **Name**: Beschreibender Name für das Mapping
   - **Event Typ**: Art des MIDI-Events (Note On, Control Change, etc.)
   - **Event-Parameter**: Abhängig vom Event-Typ (Note, Controller, Velocity, etc.)
   - **Aktion Typ**: Art der Systemaktion (Volume, App Start, etc.)
   - **Aktions-Parameter**: Abhängig vom Aktionstyp
3. Klicken Sie auf "Speichern"

#### Mapping bearbeiten
1. Klicken Sie auf "Bearbeiten" neben dem gewünschten Mapping
2. Ändern Sie die gewünschten Werte im Modal-Dialog
3. Klicken Sie auf "Speichern"

#### Mapping löschen
1. Klicken Sie auf "Löschen" neben dem gewünschten Mapping
2. Bestätigen Sie die Löschung im Bestätigungsdialog

#### Konfiguration speichern
- Klicken Sie auf "Speichern" in der Toolbar
- Die Konfiguration wird automatisch in `config.json` gespeichert
- Eine Bestätigungsmeldung wird angezeigt

#### Aktualisieren
- Klicken Sie auf "Aktualisieren" um die Mapping-Liste neu zu laden
- Nützlich wenn die Konfiguration von außen geändert wurde

### Controller-Tab

#### Controller erkennen
1. Klicken Sie auf "Controller erkennen"
2. Das System scannt automatisch alle verfügbaren MIDI-Ports
3. Erkannte Controller werden mit Details angezeigt

#### Controller-Informationen
Jeder erkannte Controller zeigt:
- **Name und Hersteller**: Identifikation des Geräts
- **Typ**: Keyboard, Pad, Knob, Slider, Mixer, DJ, etc.
- **Status**: Verbunden/Getrennt
- **Fähigkeiten**: Anzahl Tasten, Drehregler, Pads, etc.
- **Spezielle Features**: Display, Transport-Controls, etc.

#### Controller-Einstellungen
1. Klicken Sie auf "Einstellungen" bei einem Controller
2. Konfigurieren Sie:
   - **Standard-Kanal**: MIDI-Kanal (0-15)
   - **Velocity-sensitiv**: Velocity-Informationen verwenden
   - **Druck-sensitiv**: Aftertouch/Expression verwenden
   - **Automatisches Mapping**: Vorgeschlagene Mappings automatisch anwenden
3. Klicken Sie auf "Speichern"

#### Vorgeschlagene Mappings
1. Klicken Sie auf "Vorschläge" bei einem Controller
2. Das System zeigt intelligente Mapping-Vorschläge basierend auf:
   - Controller-Typ (Keyboard, Pad, etc.)
   - Verfügbare Fähigkeiten
   - Häufige Anwendungsfälle
3. Klicken Sie auf "Mapping hinzufügen" um einen Vorschlag zu übernehmen
4. Das Mapping-Formular wird automatisch ausgefüllt

## Unterstützte Controller

### Akai Professional
- **MPK Mini**: 25 Tasten, 8 Drehregler, 8 Pads
- **MPK249**: 49 Tasten, 8 Drehregler, 16 Pads, 8 Schieberegler
- **MPX**: 16 Pads mit Display

### Native Instruments
- **Traktor Kontrol**: DJ-Controller mit Transport-Controls
- **Maschine**: Drum Machine mit 16 Pads

### Behringer
- **X32**: Digitaler Mixer mit 32 Fadern
- **XR18**: Kompakter Mixer mit 18 Fadern

### Arturia
- **KeyLab**: Keyboard mit 61 Tasten, 9 Drehregler/Schieberegler
- **BeatStep**: Drum Machine mit 16 Pads und 16 Drehreglern

### Novation
- **LaunchKey Mini**: 25 Tasten, 8 Drehregler, 16 Pads
- **LaunchKey 49/61**: Erweiterte Versionen
- **LaunchPad**: 64 Pads mit Display

### M-Audio
- **Oxygen**: Keyboard mit 49 Tasten, 8 Drehregler, 9 Schieberegler

### Korg
- **Nano**: Kompakte Controller-Serie

### Roland
- **A-Serie**: Professionelle Keyboards mit Modulationsrad

### Yamaha
- **Motif**: Synthesizer mit 88 Tasten und Display

### Generic Controller
- Automatische Erkennung unbekannter MIDI-Controller
- Basis-Funktionalität basierend auf Port-Namen

## Action-Typen und Parameter

### Volume
- **Parameter**: `up`, `down`, `mute`, `unmute`
- **Beispiel**: `up` für Lautstärke erhöhen

### App Start
- **Parameter**: Pfad zur ausführbaren Datei
- **Beispiel**: `notepad.exe`, `/usr/bin/firefox`

### Key Combination
- **Parameter**: Tastenkombination
- **Beispiel**: `Ctrl+Alt+Del`, `Ctrl+C`

### Audio Source
- **Parameter**: Name der Audioquelle
- **Beispiel**: `Speakers`, `Headphones`

## Technische Details

### Architektur

Die GUI verwendet eine Web-basierte Architektur:

- **Backend**: Go HTTP-Server mit REST-API
- **Frontend**: HTML5, CSS3, JavaScript (Vanilla JS)
- **Kommunikation**: JSON über HTTP-API

### API-Endpunkte

#### Mapping-Management
- `GET /api/mappings` - Alle Mappings abrufen
- `POST /api/save` - Konfiguration speichern
- `POST /api/add` - Neues Mapping hinzufügen
- `POST /api/edit` - Mapping bearbeiten
- `POST /api/delete` - Mapping löschen

#### Controller-Erkennung
- `GET /api/controllers` - Alle erkannten Controller abrufen
- `POST /api/discover` - Controller-Erkennung starten
- `GET /api/controller/{id}` - Controller-Details abrufen
- `PUT /api/controller/{id}` - Controller-Einstellungen aktualisieren
- `GET /api/suggestions/{id}` - Vorgeschlagene Mappings abrufen

### Dateistruktur

```
cmd/mididaemon-gui/
├── main.go                    # Hauptdatei der GUI-Anwendung
└── templates/
    └── index.html            # HTML-Template für die Benutzeroberfläche
```

### Controller-Erkennung

Das System verwendet intelligente Pattern-Matching-Algorithmen:

1. **Port-Scanning**: Alle verfügbaren MIDI-Ports werden gescannt
2. **Pattern-Matching**: Port-Namen werden gegen bekannte Controller-Patterns geprüft
3. **Capability-Detection**: Controller-Fähigkeiten werden basierend auf Modell erkannt
4. **Settings-Management**: Gerätespezifische Einstellungen werden verwaltet

### Vorgeschlagene Mappings

Das System generiert intelligente Mapping-Vorschläge basierend auf:

- **Controller-Typ**: Keyboard, Pad, Knob, etc.
- **Verfügbare Fähigkeiten**: Anzahl Tasten, Drehregler, etc.
- **Häufige Anwendungsfälle**: Volume-Control, Transport, etc.
- **Priorität**: Wichtige Mappings werden zuerst vorgeschlagen

## Fehlerbehebung

### GUI startet nicht

- Überprüfen Sie, ob Port 8080 verfügbar ist
- Stellen Sie sicher, dass Go korrekt installiert ist
- Führen Sie `go mod tidy` aus

### Browser kann nicht verbinden

- Überprüfen Sie die Firewall-Einstellungen
- Stellen Sie sicher, dass der Server läuft
- Prüfen Sie die Konsolen-Ausgabe auf Fehler

### Controller werden nicht erkannt

- Überprüfen Sie die MIDI-Verbindung
- Stellen Sie sicher, dass der Controller eingeschaltet ist
- Prüfen Sie die Treiber-Installation
- Verwenden Sie "Controller erkennen" erneut

### Mappings funktionieren nicht

- Überprüfen Sie die Konfigurationsdatei
- Prüfen Sie die Logs auf Fehler
- Testen Sie die MIDI-Verbindung separat

### Performance-Probleme

- Reduzieren Sie die Anzahl der Mappings
- Deaktivieren Sie nicht verwendete Mappings
- Überprüfen Sie die System-Ressourcen

## Erweiterte Funktionen

### Automatisches Mapping

Bei aktiviertem "Automatisches Mapping" werden neue Controller automatisch mit sinnvollen Standard-Mappings konfiguriert.

### Custom Mappings

Controller-spezifische benutzerdefinierte Mappings können in den Controller-Einstellungen konfiguriert werden.

### Multi-Controller-Support

Das System unterstützt mehrere gleichzeitig angeschlossene Controller und verwaltet sie unabhängig voneinander.

### Backup und Restore

Die Konfiguration wird automatisch in `config.json` gespeichert und kann einfach gesichert/restauriert werden.

## Entwicklung

### Lokale Entwicklung

```bash
# Entwicklungsserver starten
cd cmd/mididaemon-gui
go run main.go

# Mit Hot-Reload (falls verfügbar)
air
```

### Debugging

- Browser-Entwicklertools für Frontend-Debugging
- Go-Logs für Backend-Debugging
- MIDI-Monitoring für Controller-Tests

### Erweiterungen

Die GUI ist modular aufgebaut und kann einfach erweitert werden:

- Neue Controller-Typen hinzufügen
- Zusätzliche Action-Typen implementieren
- UI-Komponenten erweitern
- API-Endpunkte hinzufügen 