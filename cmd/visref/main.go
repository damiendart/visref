// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"flag"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/sqlite"
)

type application struct {
	config         config
	logger         *slog.Logger
	ItemRepository library.ItemRepository
	templateCache  TemplateCache
}

type config struct {
	baseURL  string
	database string
	httpPort int
	mediaDir string
}

var cfg config

func init() {
	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.StringVar(&cfg.database, "main-database-path", "visref.db", "relative path to main database")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.mediaDir, "media-dir", "media", "relative path to directory for storing media items")

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
	err := os.MkdirAll(cfg.mediaDir, os.ModePerm)
	if err != nil {
		return err
	}

	templateCache, err := NewTemplateCache()
	if err != nil {
		return err
	}

	mainDatabase := sqlite.NewMainDB(cfg.database, logger)
	if err = mainDatabase.Open(); err != nil {
		return err
	}

	app := &application{
		config:         cfg,
		logger:         logger,
		ItemRepository: sqlite.NewItemRepository(&mainDatabase.DB, cfg.mediaDir),
		templateCache:  templateCache,
	}

	return app.serveHTTP()
}
