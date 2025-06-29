package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Xcruser/MidiDaemon/pkg/utils"
)

var (
	Version   = "dev"
	BuildTime = ""
	GitCommit = ""
)

func printVersion() {
	fmt.Printf("MidiDaemon GUI %s (%s, %s)\n", Version, GitCommit, BuildTime)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP Listen-Adresse")
	verbose := flag.Bool("verbose", false, "Ausführliche Log-Ausgabe")
	debug := flag.Bool("debug", false, "Aktiviere Debug-Modus")
	logLevel := flag.String("log-level", "", "Log-Level (debug, info, warn, error)")
	showVersion := flag.Bool("version", false, "Versionsinformationen anzeigen")

	flag.Parse()

	if *showVersion {
		printVersion()
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "MidiDaemon GUI")
	})

	srv := &http.Server{Addr: *addr, Handler: mux}

	go func() {
		logger.Info("Starte GUI", "addr", *addr, "version", Version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP-Server Fehler", "error", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	logger.Info("Beende GUI ...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error("Fehler beim Herunterfahren", "error", err)
	}
}
