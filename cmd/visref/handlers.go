// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"errors"
	"fmt"
	"github.com/damiendart/visref/internal/library"
	"net/http"

	"github.com/damiendart/visref/internal/httputil"
)

func (app *application) itemsAddHandler() http.Handler {
	return httputil.ComposableHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) http.Handler {
			m := library.Item{Title: "Test"}

			err := app.ItemRepository.Create(r.Context(), &m)
			if err != nil {
				return httputil.Error(
					errors.New("oh no"),
					http.StatusInternalServerError,
				)
			}

			return httputil.Text(
				fmt.Sprintf("test library item %q created", m.ID),
			)
		},
	)
}

func (app *application) itemsIndexHandler() http.Handler {
	return httputil.ComposableHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) http.Handler {
			t, ok := app.templateCache["index.gohtml"]
			if !ok {
				return httputil.Error(
					errors.New("template index.gohtml does not exist"),
					http.StatusInternalServerError,
				)
			}

			return httputil.Template(*t, "base", nil)
		},
	)
}

func (app *application) itemsShowHandler() http.Handler {
	return httputil.ComposableHandlerFunc(
		func(w http.ResponseWriter, r *http.Request) http.Handler {
			return httputil.Text(fmt.Sprintf("itemsShow: %v", r.PathValue("id")))
		},
	)
}

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
