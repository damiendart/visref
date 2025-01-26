// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/damiendart/visref/cmd/visref/resources"
	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/sqlite"
)

type application struct {
	config         config
	logger         *slog.Logger
	LibraryService *library.Service
	templateCache  resources.TemplateCache
}

type config struct {
	baseURL  string
	dataDir  string
	httpPort int
}

var cfg config

func init() {
	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.StringVar(&cfg.dataDir, "data-dir", "data", "relative path to directory for storing application data")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")

	flag.Parse()
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	mediaDir := filepath.Join(cfg.dataDir, "media")

	err := os.MkdirAll(mediaDir, 0700)
	if err != nil {
		return err
	}

	templateCache, err := resources.NewTemplateCache()
	if err != nil {
		return err
	}

	mainDatabase := sqlite.NewMainDB(filepath.Join(cfg.dataDir, "main.db"), logger)
	if err = mainDatabase.Open(); err != nil {
		return err
	}

	app := &application{
		config:         cfg,
		logger:         logger,
		LibraryService: library.NewService(&mainDatabase.DB, mediaDir),
		templateCache:  templateCache,
	}

	return app.serveHTTP()
}
