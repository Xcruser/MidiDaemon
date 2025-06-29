# MidiDaemon GUI

Die MidiDaemon GUI ist eine benutzerfreundliche Web-basierte Oberfläche zur Konfiguration von MIDI-Mappings.

## Features

- **Mapping-Übersicht**: Anzeige aller konfigurierten MIDI-Mappings in einer übersichtlichen Tabelle
- **Mapping-Editor**: Einfaches Hinzufügen, Bearbeiten und Löschen von Mappings über ein Modal-Dialog
- **Live-Validierung**: Sofortige Überprüfung der Eingaben auf Gültigkeit
- **Konfigurationsspeicherung**: Direktes Speichern der Änderungen in die config.json
- **Plattformübergreifend**: Funktioniert auf allen Plattformen über den Webbrowser
- **Responsive Design**: Moderne, benutzerfreundliche Oberfläche

## Installation

### Voraussetzungen

- Go 1.24.4 oder höher
- Webbrowser (Chrome, Firefox, Safari, Edge)

### Build

```bash
# GUI für aktuelle Plattform bauen
make build-gui

# GUI für alle Plattformen bauen
make build-gui-all

# GUI direkt ausführen
make run-gui
```

### Manueller Build

```bash
# Abhängigkeiten installieren
go mod tidy

# GUI bauen
go build -o mididaemon-gui cmd/mididaemon-gui/main.go

# GUI ausführen
./mididaemon-gui
```

## Verwendung

### Starten der GUI

```bash
# Über Makefile
make run-gui

# Direkt
go run cmd/mididaemon-gui/main.go

# Gebautes Binary
./mididaemon-gui
```

Die GUI startet einen lokalen Web-Server auf Port 8080. Öffnen Sie Ihren Browser und navigieren Sie zu:

```
http://localhost:8080
```

### Mapping hinzufügen

1. Klicken Sie auf "Mapping hinzufügen" in der Toolbar
2. Füllen Sie das Formular aus:
   - **Name**: Beschreibender Name für das Mapping
   - **Event Typ**: Art des MIDI-Events (note_on, note_off, control_change, program_change)
   - **MIDI Note**: MIDI-Note (0-127)
   - **Action Typ**: Art der Aktion (volume, app_start, key_combination, audio_source)
   - **Parameter**: Parameter für die Aktion
   - **Aktiviert**: Checkbox zum Aktivieren/Deaktivieren des Mappings
3. Klicken Sie auf "Speichern"

### Mapping bearbeiten

1. Klicken Sie auf "Bearbeiten" neben dem gewünschten Mapping
2. Ändern Sie die gewünschten Werte im Modal-Dialog
3. Klicken Sie auf "Speichern"

### Mapping löschen

1. Klicken Sie auf "Löschen" neben dem gewünschten Mapping
2. Bestätigen Sie die Löschung im Bestätigungsdialog

### Konfiguration speichern

- Klicken Sie auf "Speichern" in der Toolbar
- Die Konfiguration wird automatisch in `config.json` gespeichert
- Eine Bestätigungsmeldung wird angezeigt

### Aktualisieren

- Klicken Sie auf "Aktualisieren" um die Mapping-Liste neu zu laden
- Nützlich wenn die Konfiguration von außen geändert wurde

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

- `GET /` - Hauptseite
- `GET /api/mappings` - Alle Mappings abrufen
- `POST /api/save` - Konfiguration speichern
- `POST /api/add` - Neues Mapping hinzufügen
- `POST /api/edit` - Mapping bearbeiten
- `POST /api/delete` - Mapping löschen

### Dateistruktur

```
cmd/mididaemon-gui/
├── main.go                    # Hauptdatei der GUI-Anwendung
└── templates/
    └── index.html            # HTML-Template für die Benutzeroberfläche
```

## Fehlerbehebung

### GUI startet nicht

- Überprüfen Sie, ob Port 8080 verfügbar ist
- Stellen Sie sicher, dass Go korrekt installiert ist
- Führen Sie `go mod tidy` aus

### Browser kann nicht verbinden

- Überprüfen Sie, ob der Server läuft: `http://localhost:8080`
- Prüfen Sie Firewall-Einstellungen
- Versuchen Sie einen anderen Port

### Mappings werden nicht gespeichert

- Überprüfen Sie die Schreibrechte im Projektverzeichnis
- Stellen Sie sicher, dass `config.json` existiert und beschreibbar ist
- Prüfen Sie die Browser-Konsole auf JavaScript-Fehler

### Template-Datei nicht gefunden

- Stellen Sie sicher, dass `cmd/mididaemon-gui/templates/index.html` existiert
- Überprüfen Sie den Arbeitsverzeichnis-Pfad

## Entwicklung

### Erweitern der GUI

Um neue Features hinzuzufügen:

1. Erweitern Sie die API-Endpunkte in `main.go`
2. Fügen Sie neue UI-Komponenten in `templates/index.html` hinzu
3. Implementieren Sie die JavaScript-Logik
4. Aktualisieren Sie die Dokumentation

### Styling anpassen

Die GUI verwendet CSS3 für das Styling. Änderungen können in der `<style>`-Sektion von `templates/index.html` vorgenommen werden.

### Neue Action-Typen

Um neue Action-Typen hinzuzufügen:

1. Erweitern Sie die Konfigurationsstruktur in `internal/config/config.go`
2. Fügen Sie den neuen Typ zur Action-Typ-Auswahl in der GUI hinzu
3. Implementieren Sie die Parameter-Logik im JavaScript
4. Aktualisieren Sie die Validierung

## Lizenz

Die GUI steht unter der gleichen Lizenz wie das Hauptprojekt (MIT License). 