// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import "net/http"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
