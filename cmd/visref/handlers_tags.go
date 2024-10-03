// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"github.com/damiendart/visref/internal/httputil"
	"net/http"
)

func (app *application) tagsAddHandler() http.Handler {
	return httputil.Text("tagsAdd")
}

func (app *application) tagsIndexHandler() http.Handler {
	return httputil.Text("tagsIndex")
}

func (app *application) tagsShowHandler() http.Handler {
	return httputil.ComposableHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) http.Handler {
			return httputil.Text(fmt.Sprintf("tagsShow: %v", r.PathValue("tag")))
		},
	)
}
