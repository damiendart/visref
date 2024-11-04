// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"

	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/validator"
)

type itemAddForm struct {
	Title       string
	Description string

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
		Title:       r.PostForm.Get("title"),
		Description: r.PostFormValue("description"),
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

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")

	if !form.Valid() {
		app.render(w, http.StatusUnprocessableEntity, "items_add.gohtml", form)
		return
	}

	m := library.Item{
		Title:            form.Title,
		Description:      form.Description,
		MimeType:         filetype,
		OriginalFilename: header.Filename,
	}

	err = app.ItemRepository.Create(r.Context(), &m, file)
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

	item, err := app.ItemRepository.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "itemsShow: %v", item)
}
