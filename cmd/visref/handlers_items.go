// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/damiendart/visref/internal/httputil"
	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/validator"
)

// ItemAddForm holds form values and errors for the form on the "Add a
// Library Item" page.
type ItemAddForm struct {
	AlternativeText string
	Source          string
	Description     string
	validator.FormValidator
}

func (app *application) itemsAddHandler() httputil.ChainableHandler {
	return app.withTemplate("items_add.gohtml", nil, http.StatusOK)
}

func (app *application) itemsAddPostHandler() httputil.ChainableHandler {
	return func(w http.ResponseWriter, r *http.Request) httputil.ChainableHandler {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*100)

		err := r.ParseMultipartForm(1024 * 1024 * 10)
		if err != nil {
			return app.withError("itemsAddPost: %w: %w", err, errBadRequest)
		}

		form := ItemAddForm{
			AlternativeText: r.PostFormValue("alternative_text"),
			Source:          r.PostFormValue("source"),
			Description:     r.PostFormValue("description"),
		}

		file, header, err := r.FormFile("media")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				form.AddError("media", "Provide a media file")
			} else {
				return app.withError("itemsAddPost: %w: %w", err, errBadRequest)
			}
		}

		if form.HasErrors() {
			return app.withTemplate("items_add.gohtml", &form, http.StatusUnprocessableEntity)
		}

		form.Check(header.Size <= 1024*1024*10, "media", "The media file must be 10 MB or smaller")

		defer file.Close()

		buf := make([]byte, 512)
		_, err = file.Read(buf)
		if err != nil {
			return app.withError("itemsAddPost: %w", err)
		}

		filetype := http.DetectContentType(buf)
		form.Check(
			filetype == "image/jpeg" || filetype == "image/png",
			"media",
			"The media file must be a supported file type",
		)

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return app.withError("itemsAddPost: %w", err)
		}

		if form.HasErrors() {
			return app.withTemplate("items_add.gohtml", &form, http.StatusUnprocessableEntity)
		}

		m := library.Item{
			AlternativeText:  form.AlternativeText,
			Source:           form.Source,
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
