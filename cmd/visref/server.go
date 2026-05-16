// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"context"
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

	app.logger.LogAttrs(
		context.TODO(),
		slog.LevelInfo,
		"starting server",
		slog.GroupAttrs(
			"server",
			slog.String("addr", srv.Addr),
		),
	)

	return srv.ListenAndServe()
}
