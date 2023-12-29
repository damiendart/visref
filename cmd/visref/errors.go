package main

import (
	"net/http"
)

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404", http.StatusNotFound)
}
