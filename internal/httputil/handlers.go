// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package httputil

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// The ComposableHandlerFunc type is an adapter to allow the use of
// handler functions that return other handlers.
type ComposableHandlerFunc func(http.ResponseWriter, *http.Request) http.Handler

// ServeHTTP makes ComposableHandlerFunc implement the http.Handler interface.
func (f ComposableHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := f(w, r); handler != nil {
		handler.ServeHTTP(w, r)
	}
}

// Error returns a http.HandlerFunc that uses http.Error to output the
// given error message and HTTP status code to the client.
func Error(err error, code int) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), code)
		},
	)
}

// Text returns a http.HandlerFunc that outputs the given string to the client.
func Text(v any) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, v)
		},
	)
}

// Template returns a ComposableHandlerFunc that renders the given
// template.Template and outputs it to the client. The output is
// buffered beforehand to prevent partial output if an error occurs.
func Template(t template.Template, name string, p any) ComposableHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) http.Handler {
		var b []byte
		buffer := bytes.NewBuffer(b)

		err := t.ExecuteTemplate(buffer, name, p)
		if err != nil {
			return Error(err, http.StatusInternalServerError)
		}

		_, err = buffer.WriteTo(w)
		if err != nil {
			return Error(err, http.StatusInternalServerError)
		}

		return nil
	}
}
