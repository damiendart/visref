// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"

	"github.com/damiendart/visref/internal/httputil"
)

func (app *application) tagsAddHandler() httputil.ChainableHandler {
	return app.withText("tagsAdd", http.StatusOK)
}

func (app *application) tagsIndexHandler() httputil.ChainableHandler {
	return app.withText("tagsIndex", http.StatusOK)
}

func (app *application) tagsShowHandler() httputil.ChainableHandler {
	return app.withText("tagsShow", http.StatusOK)
}
