// Copyright (C) Damien Dart, <damiendart@pobox.com>.
// This file is distributed under the MIT licence. For more information,
// please refer to the accompanying "LICENCE" file.

package validator

import "strings"

// Validator ...
type Validator struct {
	Errors map[string]string
}

// CheckField ...
func (v *Validator) CheckField(ok bool, key string, message string) {
	if !ok {
		if v.Errors == nil {
			v.Errors = make(map[string]string)
		}

		_, exists := v.Errors[key]
		if !exists {
			v.Errors[key] = message
		}
	}
}

// Valid ...
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// NotBlank ...
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}
