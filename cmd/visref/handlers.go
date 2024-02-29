// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"net/http"
)

func (app *application) mediaAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "mediaAdd")
}

func (app *application) mediaIndex(w http.ResponseWriter, r *http.Request) {
	app.views.renderIndex(w)
}

func (app *application) mediaShow(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "mediaShow: %v", r.PathValue("id"))
}

func (app *application) tagsAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "tagsAdd")
}

func (app *application) tagsIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "tagsIndex")
}

func (app *application) tagsShow(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "tagsShow: %v", r.PathValue("tag"))
}
