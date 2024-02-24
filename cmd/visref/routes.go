// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /media", http.RedirectHandler("/", http.StatusFound))

	mux.HandleFunc("GET /{$}", app.mediaIndex)
	mux.HandleFunc("GET /media/add", app.mediaAdd)
	mux.HandleFunc("GET /media/{id}", app.mediaShow)

	mux.HandleFunc("GET /tags", app.tagsIndex)
	mux.HandleFunc("GET /tags/add", app.tagsAdd)
	mux.HandleFunc("GET /tags/{tag}", app.tagsShow)

	return mux
}
