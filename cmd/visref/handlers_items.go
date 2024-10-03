// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package main

import (
	"fmt"
	"github.com/damiendart/visref/internal/library"
	"github.com/damiendart/visref/internal/validator"
	"github.com/google/uuid"
	"net/http"
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
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	form := itemAddForm{
		Title:       r.PostForm.Get("title"),
		Description: r.PostFormValue("description"),
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")

	if !form.Valid() {
		app.render(w, http.StatusUnprocessableEntity, "items_add.gohtml", form)
		return
	}

	m := library.Item{
		Title:       form.Title,
		Description: form.Description,
	}

	err = app.ItemRepository.Create(r.Context(), &m)
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
