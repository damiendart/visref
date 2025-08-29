// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/damiendart/visref/internal/httputil"
	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/validator"
)

type itemAddForm struct {
	AlternativeText string
	Description     string
	validator.Validator
}

func (app *application) itemsAddHandler() httputil.ChainableHandler {
	return app.withTemplate("items_add.gohtml", nil, http.StatusOK)
}

func (app *application) itemsAddPostHandler() httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*10)

		err := r.ParseMultipartForm(1024 * 1024 * 10)
		if err != nil {
			return app.withError("itemsAddPost: %w: %w", err, errBadRequest)
		}

		form := itemAddForm{
			AlternativeText: r.PostForm.Get("alternative_text"),
			Description:     r.PostFormValue("description"),
		}

		file, header, err := r.FormFile("media")
		if err != nil {
			return app.withError("itemsAddPost: %w: %w", err, errBadRequest)
		}

		defer file.Close()

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		if err != nil {
			return app.withError("itemsAddPost: %w", err)
		}

		filetype := http.DetectContentType(buf)
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return app.withError("itemsAddPost: %w", err)
		}

		if form.HasErrors() {
			return app.withTemplate("items_add.gohtml", form, http.StatusUnprocessableEntity)
		}

		m := library.Item{
			AlternativeText:  form.AlternativeText,
			Description:      form.Description,
			MimeType:         filetype,
			OriginalFilename: header.Filename,
		}

		err = app.LibraryService.CreateItem(r.Context(), &m, file)
		if err != nil {
			return app.withError("%w", err)
		}

		return app.withRedirect(fmt.Sprintf("/items/%s", m.ID), http.StatusSeeOther)
	}
}

func (app *application) itemsIndexHandler() httputil.ChainableHandler {
	return app.withTemplate("index.gohtml", nil, http.StatusOK)
}

func (app *application) itemsShowHandler() httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			return app.withError("itemShow: %w", errNotFound)
		}

		item, err := app.LibraryService.GetItemByID(r.Context(), id)
		if err != nil {
			return app.withError("itemShow: %w", errNotFound)
		}

		if r.URL.Query().Get("download") == "1" {
			f, err := app.LibraryService.GetOriginalFileByItem(item)
			if err != nil {
				return app.withError("itemShow: %w", err)
			}

			return app.withContent(item.OriginalFilename, item.CreatedAt, f)
		}

		return app.withText(fmt.Sprintf("itemsShow: %v", item), http.StatusOK)
	}
}
