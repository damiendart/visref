// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"log/slog"
	"net/http"
)

// DefaultHeaders is an HTTP middleware function that adds a few common
// HTTP headers that apply to all requests.
func DefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("X-Frame-Options", "deny")

			next.ServeHTTP(w, r)
		},
	)
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			app.logger.Info(
				"access",
				slog.Group("request", "method", r.Method, "proto", r.Proto, "url", r.URL.String()),
			)

			next.ServeHTTP(w, r)
		},
	)
}
