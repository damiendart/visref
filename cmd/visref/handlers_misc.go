// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"

	"github.com/damiendart/visref/internal/httputil"
)

func (app *application) helpShowHandler() httputil.ChainableHandler {
	return app.withTemplate("help_show.gohtml", nil, http.StatusOK)
}
