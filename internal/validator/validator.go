// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package validator

// FormValidator contains errors messages for forms and their fields.
type FormValidator struct {
	Errors map[string]string
}

// AddError adds the field error message to the Validator.
func (v *FormValidator) AddError(key string, message string) {
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}

	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

// Check adds the field error message to the Validator if the validation
// check fails.
func (v *FormValidator) Check(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// HasErrors reports whether the Validator contains any error messages.
func (v *FormValidator) HasErrors() bool {
	return len(v.Errors) != 0
}
