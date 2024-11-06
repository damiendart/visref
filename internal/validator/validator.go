// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package validator

import "strings"

// Validator contains errors messages for forms and their fields.
type Validator struct {
	Errors      []string
	FieldErrors map[string]string
}

// AddError adds the form error message to the Validator.
func (v *Validator) AddError(message string) {
	if v.Errors == nil {
		v.Errors = []string{}
	}

	v.Errors = append(v.Errors, message)
}

// AddFieldError adds the field error message to the Validator.
func (v *Validator) AddFieldError(key string, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	_, exists := v.FieldErrors[key]
	if !exists {
		v.FieldErrors[key] = message
	}
}

// Check adds the form error message to the Validator if the validation
// check fails.
func (v *Validator) Check(ok bool, message string) {
	if !ok {
		v.AddError(message)
	}
}

// CheckField adds the field error message to the Validator if the
// validation check fails.
func (v *Validator) CheckField(ok bool, key string, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// HasErrors reports whether the Validator contains any error messages.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) != 0 || len(v.FieldErrors) != 0
}

// NotBlank reports whether the string s is not blank.
func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}
