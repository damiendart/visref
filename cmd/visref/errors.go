// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"net/http"
)

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404", http.StatusNotFound)
}
