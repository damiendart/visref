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

	mux.HandleFunc("GET /{$}", app.itemsIndexHandler)
	mux.Handle("GET /assets/", http.FileServer(http.FS(resources)))
	mux.Handle("GET /items", http.RedirectHandler("/", http.StatusFound))
	mux.HandleFunc("GET /items/{id}", app.itemsShowHandler)
	mux.HandleFunc("GET /tags", app.tagsIndexHandler)
	mux.HandleFunc("GET /tags/{tag}", app.tagsShowHandler)

	mux.Group(
		func(m httputil.Router) {
			m.HandleFunc("GET /items/add", app.itemsAddHandler)
			m.HandleFunc("POST /items/add", app.itemsAddPostHandler)
			m.HandleFunc("GET /tags/add", app.tagsAddHandler)
		},
	)

	return mux
}
