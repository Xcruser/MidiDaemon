# Multi-stage build für MidiDaemon
FROM golang:1.21-alpine AS builder

# Arbeitsverzeichnis setzen
WORKDIR /app

# Abhängigkeiten installieren
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren
COPY . .

# Binary bauen
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mididaemon ./cmd/mididaemon

# Runtime-Image
FROM alpine:latest

# Runtime-Abhängigkeiten installieren
RUN apk --no-cache add ca-certificates tzdata

# Arbeitsverzeichnis erstellen
WORKDIR /root/

# Binary aus dem Builder-Image kopieren
COPY --from=builder /app/mididaemon .

# Konfigurationsdatei kopieren
COPY --from=builder /app/config.json .

# Ports exponieren (falls benötigt)
EXPOSE 8080

# Umgebungsvariablen setzen
ENV CONFIG_FILE=config.json

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ps aux | grep mididaemon || exit 1

# Binary ausführen
ENTRYPOINT ["./mididaemon"]
CMD ["-config", "config.json"] 