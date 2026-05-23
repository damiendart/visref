// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
var version string

func init() {
	var printVersion bool

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.StringVar(&cfg.dataDir, "data-dir", "data", "path to directory for storing application data")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.BoolVar(&printVersion, "version", false, "print application version and exit")

	buildInfo, _ := debug.ReadBuildInfo()

	if buildInfo.Main.Version != "" {
		version = buildInfo.Main.Version
	} else {
		version = "unknown"
	}

	flag.Parse()

	if printVersion {
		fmt.Println(version)
		os.Exit(0)
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	if err := run(logger); err != nil {
		trace := string(debug.Stack())
		logger.LogAttrs(
			context.Background(),
			slog.LevelError,
			err.Error(),
			slog.String("trace", trace),
		)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	dataDir, err := filepath.Abs(cfg.dataDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dataDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(dataDir, 0700); err != nil {
				return err
			}

			logger.LogAttrs(
				context.Background(),
				slog.LevelInfo,
				"created data directory",
				slog.String("data_dir", dataDir),
			)
		} else {
			return err
		}
	}

	dataRoot, err := os.OpenRoot(dataDir)
	if err != nil {
		return err
	}
	defer dataRoot.Close()

	err = dataRoot.MkdirAll("media", 0700)
	if err != nil {
		return err
	}

	mediaRoot, err := dataRoot.OpenRoot("media")
	if err != nil {
		return err
	}
	defer mediaRoot.Close()

	templateCache, err := resources.NewTemplateCache()
	if err != nil {
		return err
	}

	mainDatabase := sqlite.NewMainDB(filepath.Join(dataRoot.Name(), "main.db"), logger)
	if err = mainDatabase.Open(); err != nil {
		return err
	}

	logger.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"starting application",
		slog.GroupAttrs(
			"application",
			slog.String("version", version),
			slog.String("data_dir", dataDir),
		),
	)

	app := &application{
		config:         cfg,
		logger:         logger,
		LibraryService: library.NewService(&mainDatabase.DB, mediaRoot),
		templateCache:  templateCache,
	}

	return app.serveHTTP()
}
