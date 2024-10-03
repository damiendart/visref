// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"net/http"
)

func (app *application) tagsAddHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "tagsAdd")
}

func (app *application) tagsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "tagsIndex")
}

func (app *application) tagsShowHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "tagsShow: %v", r.PathValue("tag"))
}
