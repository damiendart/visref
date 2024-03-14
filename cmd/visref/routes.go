// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"

	"github.com/damiendart/visref/internal/httputil"
)

func (app *application) routes() http.Handler {
	mux := httputil.NewRouter()

	mux.Use(DefaultHeaders)

	mux.Handle("GET /media", http.RedirectHandler("/", http.StatusFound))

	mux.Handle("GET /{$}", app.mediaIndexHandler())
	mux.Handle("GET /media/add", app.mediaAddHandler())
	mux.Handle("GET /media/{id}", app.mediaShowHandler())

	mux.Handle("GET /tags", app.tagsIndexHandler())
	mux.Handle("GET /tags/add", app.tagsAddHandler())
	mux.Handle("GET /tags/{tag}", app.tagsShowHandler())

	return mux
}
