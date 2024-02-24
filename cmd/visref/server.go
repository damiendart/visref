// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

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
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// Add support for spoofing unsupported HTML form
				// actions ("PUT", "PATCH", and "DELETE") with a hidden
				// "_method" input field.
				if r.Method == http.MethodPost {
					switch m := r.PostFormValue("_method"); m {
					case http.MethodDelete, http.MethodPatch, http.MethodPut:
						r.Method = m
					}
				}

				app.routes().ServeHTTP(w, r)
			},
		),
	}

	app.logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	return srv.ListenAndServe()
}
