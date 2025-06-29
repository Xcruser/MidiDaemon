# Makefile für MidiDaemon
# Ein plattformübergreifender MIDI-Controller-Daemon

# Variablen
BINARY_NAME=mididaemon
GUI_BINARY_NAME=mididaemon-gui
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Go-Variablen
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=1

# Build-Flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Verzeichnisse
CMD_DIR=cmd/mididaemon
BUILD_DIR=build
DIST_DIR=dist

# Plattformen
PLATFORMS=windows linux darwin
ARCHITECTURES=amd64 arm64

# Standard-Target
.PHONY: all
all: clean build

# Abhängigkeiten installieren
.PHONY: deps
deps:
	@echo "Installiere Abhängigkeiten..."
	$(GO) mod download
	$(GO) mod tidy

# Code formatieren
.PHONY: fmt
fmt:
	@echo "Formatiere Code..."
	$(GO) fmt ./...

# Code linten
.PHONY: lint
lint:
	@echo "Linte Code..."
	golangci-lint run

# Tests ausführen
.PHONY: test
test:
	@echo "Führe Tests aus..."
	$(GO) test -v ./...

# Tests mit Coverage ausführen
.PHONY: test-coverage
test-coverage:
	@echo "Führe Tests mit Coverage aus..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage-Report erstellt: coverage.html"

# Build-Verzeichnis erstellen
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Binary für aktuelle Plattform bauen
.PHONY: build
build: $(BUILD_DIR)
	@echo "Baue MidiDaemon für $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Binary für Windows bauen
.PHONY: build-windows
build-windows:
	@echo "Baue MidiDaemon für Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./$(CMD_DIR)

# Binary für Linux bauen
.PHONY: build-linux
build-linux:
	@echo "Baue MidiDaemon für Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./$(CMD_DIR)

# Binary für macOS bauen
.PHONY: build-darwin
build-darwin:
	@echo "Baue MidiDaemon für macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin ./$(CMD_DIR)

# Alle Plattformen bauen
.PHONY: build-all
build-all: build-windows build-linux build-darwin
	@echo "Alle Binaries erstellt in $(BUILD_DIR)/"

# Statisches Binary bauen (ohne CGO)
.PHONY: build-static
build-static:
	@echo "Baue statisches Binary..."
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME)-static ./$(CMD_DIR)

# Release bauen (optimiert)
.PHONY: build-release
build-release: $(BUILD_DIR)
	@echo "Baue Release-Version..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-release ./$(CMD_DIR)

# Debug-Version bauen
.PHONY: build-debug
build-debug: $(BUILD_DIR)
	@echo "Baue Debug-Version..."
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug ./$(CMD_DIR)

# Binary installieren
.PHONY: install
install: build
	@echo "Installiere MidiDaemon..."
	$(GO) install ./$(CMD_DIR)

# Binary ausführen
.PHONY: run
run: build
	@echo "Führe MidiDaemon aus..."
	./$(BUILD_DIR)/$(BINARY_NAME) -config config.json

# Binary mit Debug-Ausgabe ausführen
.PHONY: run-debug
run-debug: build
	@echo "Führe MidiDaemon mit Debug-Ausgabe aus..."
	./$(BUILD_DIR)/$(BINARY_NAME) -config config.json -verbose

# Binary testen
.PHONY: run-test
run-test: build
	@echo "Teste MidiDaemon..."
	./$(BUILD_DIR)/$(BINARY_NAME) -version

# Release-Paket erstellen
.PHONY: release
release: build-all
	@echo "Erstelle Release-Paket..."
	mkdir -p $(DIST_DIR)
	cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/mididaemon-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz *
	@echo "Release-Paket erstellt: $(DIST_DIR)/mididaemon-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz"

# Clean
.PHONY: clean
clean:
	@echo "Räume auf..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean

# Hilfe anzeigen
.PHONY: help
help:
	@echo "MidiDaemon Makefile"
	@echo ""
	@echo "Verfügbare Targets:"
	@echo "  all              - Clean und Build"
	@echo "  deps             - Abhängigkeiten installieren"
	@echo "  fmt              - Code formatieren"
	@echo "  lint             - Code linten"
	@echo "  test             - Tests ausführen"
	@echo "  test-coverage    - Tests mit Coverage ausführen"
	@echo "  build            - Binary für aktuelle Plattform bauen"
	@echo "  build-windows    - Binary für Windows bauen"
	@echo "  build-linux      - Binary für Linux bauen"
	@echo "  build-darwin     - Binary für macOS bauen"
	@echo "  build-all        - Binaries für alle Plattformen bauen"
	@echo "  build-static     - Statisches Binary bauen"
	@echo "  build-release    - Release-Version bauen"
	@echo "  build-debug      - Debug-Version bauen"
	@echo "  install          - Binary installieren"
	@echo "  run              - Binary ausführen"
	@echo "  run-debug        - Binary mit Debug-Ausgabe ausführen"
	@echo "  run-test         - Binary testen"
	@echo "  release          - Release-Paket erstellen"
	@echo "  clean            - Aufräumen"
	@echo "  help             - Diese Hilfe anzeigen"

# Entwicklungs-Setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Entwicklungs-Setup abgeschlossen"
	@echo "Verwende 'make run' um MidiDaemon zu starten"

# Docker-Build
.PHONY: docker-build
docker-build:
	@echo "Baue Docker-Image..."
	docker build -t mididaemon:$(VERSION) .
	docker tag mididaemon:$(VERSION) mididaemon:latest

# Docker-Run
.PHONY: docker-run
docker-run:
	@echo "Führe Docker-Container aus..."
	docker run --rm -it --device=/dev/snd mididaemon:latest

# Benchmark
.PHONY: benchmark
benchmark:
	@echo "Führe Benchmarks aus..."
	$(GO) test -bench=. -benchmem ./...

# Race-Detector
.PHONY: race
race:
	@echo "Führe Tests mit Race-Detector aus..."
	$(GO) test -race ./...

# Vet
.PHONY: vet
vet:
	@echo "Führe go vet aus..."
	$(GO) vet ./...

# Sicherheits-Check
.PHONY: security
security:
	@echo "Führe Sicherheits-Checks aus..."
	gosec ./...

# Vollständiger Check
.PHONY: check
check: fmt lint vet test
	@echo "Alle Checks abgeschlossen"

# GUI Build Targets
.PHONY: build-gui build-gui-windows build-gui-linux build-gui-macos

build-gui: build-gui-$(OS)

build-gui-windows:
	@echo "Building GUI for Windows..."
	@mkdir -p build
	@go build -ldflags="-s -w" -o build/mididaemon-gui.exe cmd/mididaemon-gui/main.go

build-gui-linux:
	@echo "Building GUI for Linux..."
	@mkdir -p build
	@go build -ldflags="-s -w" -o build/mididaemon-gui cmd/mididaemon-gui/main.go

build-gui-macos:
	@echo "Building GUI for macOS..."
	@mkdir -p build
	@go build -ldflags="-s -w" -o build/mididaemon-gui cmd/mididaemon-gui/main.go

# Cross-compilation for GUI
build-gui-all: build-gui-windows build-gui-linux build-gui-macos

# Run GUI
run-gui:
	@echo "Running GUI..."
	@go run cmd/mididaemon-gui/main.go 