package main

import (
	"bytes"
	"fmt"
	"io"
)

// Views contains all the views required for the web application.
type Views struct {
	TemplateCache TemplateCache
}

// NewViews returns a new Views using the provided template cache.
func NewViews(templateCache TemplateCache) *Views {
	return &Views{
		TemplateCache: templateCache,
	}
}

func (views *Views) renderIndex(w io.Writer) error {
	return views.renderTemplate(w, "index.gohtml", nil)
}

// renderTemplate executes templates, writing the output to a buffer
// before writing it to the provided io.Writer to prevent malformed
// output if an error occurs.
func (views *Views) renderTemplate(w io.Writer, templateName string, parameters any) error {
	templateSet, ok := views.TemplateCache[templateName]
	if !ok {
		return fmt.Errorf("template %s does not exist", templateName)
	}

	var b []byte
	buffer := bytes.NewBuffer(b)

	err := templateSet.ExecuteTemplate(buffer, "base", parameters)
	if err != nil {
		return err
	}

	_, err = buffer.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}
