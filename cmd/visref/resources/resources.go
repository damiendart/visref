// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package resources

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
)

// Resources is an embedded collection of frontend assets and templates.
//
//go:embed assets templates
var Resources embed.FS

// TemplateCache is an in-memory map of parsed templates.
type TemplateCache map[string]*template.Template

// NewTemplateCache returns a new pre-populated TemplateCache.
func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	templates, err := fs.Glob(Resources, "templates/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, t := range templates {
		name := filepath.Base(t)

		patterns := []string{"templates/layouts/*.gohtml", t}

		ts, err := template.New(name).ParseFS(Resources, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
