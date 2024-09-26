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

	mux.Use(DefaultHeaders, app.logRequest)

	mux.Handle("GET /{$}", app.itemsIndexHandler())
	mux.Handle("GET /assets/", http.FileServer(http.FS(resources)))
	mux.Handle("GET /items", http.RedirectHandler("/", http.StatusFound))
	mux.Handle("GET /items/{id}", app.itemsShowHandler())
	mux.Handle("GET /tags", app.tagsIndexHandler())
	mux.Handle("GET /tags/{tag}", app.tagsShowHandler())

	mux.Group(
		func (m httputil.Router) {
			m.Handle("GET /items/add", app.itemsAddHandler())
			m.Handle("GET /tags/add", app.tagsAddHandler())
		},
	)

	return mux
}
