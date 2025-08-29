// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/damiendart/visref/internal/httputil"
)

var (
	errNotFound   = errors.New("not found")
	errBadRequest = errors.New("bad request")
)

func (app *application) withContent(name string, modtime time.Time, content io.ReadSeeker) httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		http.ServeContent(w, r, name, modtime, content)
		return nil
	}
}

func (app *application) withError(format string, args ...any) httputil.ChainableHandler {
	err := fmt.Errorf(format, args...)

	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		code := http.StatusInternalServerError

		switch {
		case errors.Is(err, errNotFound):
			return app.withText("404 Not Found", http.StatusNotFound)
		case errors.Is(err, errBadRequest):
			return app.withText("400 Bad Request", http.StatusBadRequest)
		default:
			// TODO: Handle internal server error logging
		}

		http.Error(w, err.Error(), code)

		return nil
	}
}

func (app *application) withRedirect(url string, code int) httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		http.Redirect(w, r, url, code)
		return nil
	}
}

func (app *application) withTemplate(template string, data any, code int) httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		ts, ok := app.templateCache[template]
		if !ok {
			return app.withError("template: template %q does not exist", template)
		}

		buf := new(bytes.Buffer)
		err := ts.ExecuteTemplate(buf, "base", data)
		if err != nil {
			return app.withError("template: %w", err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(code)
		_, _ = buf.WriteTo(w)

		return nil
	}
}

func (app *application) withText(message string, code int) httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		_, _ = fmt.Fprintf(w, "%s", message)

		return nil
	}
}
