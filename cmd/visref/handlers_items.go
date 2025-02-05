// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/validator"
)

type itemAddForm struct {
	AlternativeText string
	Description     string
	validator.Validator
}

func (app *application) itemsAddHandler(w http.ResponseWriter, _ *http.Request) {
	app.render(w, http.StatusOK, "items_add.gohtml", nil)
}

func (app *application) itemsAddPostHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*10)

	err := r.ParseMultipartForm(1024 * 1024 * 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	form := itemAddForm{
		AlternativeText: r.PostForm.Get("alternative_text"),
		Description:     r.PostFormValue("description"),
	}

	file, header, err := r.FormFile("media")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := http.DetectContentType(buf)
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if form.HasErrors() {
		app.render(w, http.StatusUnprocessableEntity, "items_add.gohtml", form)
		return
	}

	m := library.Item{
		AlternativeText:  form.AlternativeText,
		Description:      form.Description,
		MimeType:         filetype,
		OriginalFilename: header.Filename,
	}

	err = app.LibraryService.CreateItem(r.Context(), &m, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/items/%s", m.ID), http.StatusSeeOther)
}

func (app *application) itemsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	app.render(w, http.StatusOK, "index.gohtml", nil)
}

func (app *application) itemsShowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	item, err := app.LibraryService.GetItemByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("download") == "1" {
		f, err := app.LibraryService.GetOriginalFileByItem(item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, item.OriginalFilename, item.CreatedAt, f)
		return
	}

	fmt.Fprintf(w, "itemsShow: %v", item)
}
