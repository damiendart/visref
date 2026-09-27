// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func (app *application) serveHTTP() error {
	srv := &http.Server{
		Addr:     fmt.Sprintf(":%d", app.config.httpPort),
		ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelWarn),
		Handler:  app.routes(),
	}

	shutdownErr := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.logger.LogAttrs(
			context.Background(),
			slog.LevelInfo,
			"stopping server",
			slog.String("addr", srv.Addr),
			slog.String("signal", s.String()),
		)

		shutdownErr <- srv.Shutdown(context.TODO())
	}()

	app.logger.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"starting server",
		slog.String("addr", srv.Addr),
	)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	return nil
}
