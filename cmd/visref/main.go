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
	database       *sqlite.DB
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
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.StringVar(&cfg.database, "database-path", "visref.db", "relative path to database")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.mediaDir, "media-dir", "media", "relative path to directory for storing media items")

	flag.Parse()

	err := os.MkdirAll(cfg.mediaDir, os.ModePerm)
	if err != nil {
		return err
	}

	templateCache, err := NewTemplateCache()
	if err != nil {
		return err
	}

	database := sqlite.NewDB(cfg.database, logger)
	if err = database.Open(); err != nil {
		return err
	}

	app := &application{
		config:         cfg,
		database:       database,
		logger:         logger,
		ItemRepository: sqlite.NewItemRepository(database, cfg.mediaDir),
		templateCache:  templateCache,
	}

	return app.serveHTTP()
}
