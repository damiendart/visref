// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"flag"
	"log/slog"
	"os"
	"runtime/debug"
)

type application struct {
	config config
	logger *slog.Logger
	views  *Views
}

type config struct {
	baseURL  string
	httpPort int
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
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")

	flag.Parse()

	templateCache, err := NewTemplateCache()
	if err != nil {
		return err
	}

	app := &application{
		config: cfg,
		logger: logger,
		views:  NewViews(templateCache),
	}

	return app.serveHTTP()
}
