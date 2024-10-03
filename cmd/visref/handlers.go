// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func (app *application) render(w http.ResponseWriter, status int, template string, data any) {
	ts, ok := app.templateCache[template]
	if !ok {
		err := fmt.Errorf("the template %q does not exist", template)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}
