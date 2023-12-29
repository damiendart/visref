package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *application) serveHTTP() error {
	srv := &http.Server{
		Addr:     fmt.Sprintf(":%d", app.config.httpPort),
		ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelWarn),
		Handler:  app.routes(),
	}

	app.logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	return srv.ListenAndServe()
}
