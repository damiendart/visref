// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"

	"github.com/damiendart/visref/cmd/visref/resources"
	"github.com/damiendart/visref/internal/httputil"
)

func (app *application) routes() http.Handler {
	mux := httputil.NewRouter()

	mux.UseGlobal(DefaultHeaders, app.logRequest)

	mux.Handle("GET /{$}", app.itemsIndexHandler())
	mux.Handle("GET /assets/", http.FileServer(http.FS(resources.Resources)))
	mux.Handle("GET /items", http.RedirectHandler("/", http.StatusFound))
	mux.Handle("GET /items/{id}", app.itemsShowHandler())
	mux.Handle("GET /tags", app.tagsIndexHandler())
	mux.Handle("GET /tags/{tag}", app.tagsShowHandler())

	mux.Group(
		func(m httputil.SubRouter) {
			m.Handle("GET /items/add", app.itemsAddHandler())
			m.Handle("POST /items/add", app.itemsAddPostHandler())
			m.Handle("PATCH /items/{id}", app.itemsPatchHandler())
			m.Handle("GET /tags/add", app.tagsAddHandler())
		},
	)

	return mux
}
