package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Xcruser/MidiDaemon/internal/config"
	"github.com/Xcruser/MidiDaemon/internal/midi"
	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

var (
	Version   = "dev"
	BuildTime = ""
	GitCommit = ""
)

func printVersion() {
	fmt.Printf("MidiDaemon %s (%s, %s)\n", Version, GitCommit, BuildTime)
}

func main() {
	configPath := flag.String("config", "config.json", "Pfad zur Konfigurationsdatei")
	verbose := flag.Bool("verbose", false, "Ausführliche Log-Ausgabe")
	debug := flag.Bool("debug", false, "Aktiviere Debug-Modus")
	logLevel := flag.String("log-level", "", "Log-Level (debug, info, warn, error)")
	generateCfg := flag.Bool("generate-config", false, "Erzeugt eine Standard-Konfigurationsdatei")
	showVersion := flag.Bool("version", false, "Versionsinformationen anzeigen")

	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *generateCfg {
		if err := config.GenerateDefaultFile(*configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Fehler beim Erzeugen der Konfiguration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Konfigurationsdatei %s erstellt\n", *configPath)
		return
	}

	level := utils.LevelInfo
	if *debug {
		level = utils.LevelDebug
	}
	if *logLevel != "" {
		level = utils.ParseLogLevel(*logLevel)
	}
	logger := utils.NewLoggerWithLevel(level, *verbose)

	logger.Info("Starte MidiDaemon", "version", Version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Fehler beim Laden der Konfiguration", "error", err)
		return
	}

	handler, err := midi.NewHandler(cfg, logger)
	if err != nil {
		logger.Fatal("Fehler beim Initialisieren", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("Beende MidiDaemon ...")
		cancel()
	}()

	if err := handler.Start(ctx); err != nil {
		logger.Error("Handler beendet", "error", err)
	}

	if err := handler.Close(); err != nil {
		logger.Error("Fehler beim Schließen des Handlers", "error", err)
	}

	logger.Info("MidiDaemon beendet")
}
