package logic

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func validateText(field, value string, maxRunes int, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return badRequest("missing " + field)
	}
	if utf8.RuneCountInString(value) > maxRunes {
		return badRequest(fmt.Sprintf("%s is too long", field))
	}
	return nil
}

func validateStudentFields(studentNo, name, gender, phone, position string) error {
	checks := []struct {
		field    string
		value    string
		maxRunes int
		required bool
	}{
		{field: "studentNo", value: studentNo, maxRunes: 64, required: true},
		{field: "name", value: name, maxRunes: 64, required: true},
		{field: "gender", value: gender, maxRunes: 16},
		{field: "phone", value: phone, maxRunes: 32},
		{field: "position", value: position, maxRunes: 64},
	}
	for _, check := range checks {
		if err := validateText(check.field, check.value, check.maxRunes, check.required); err != nil {
			return err
		}
	}
	return nil
}

func validateScoreValue(score int64) error {
	if score < -1000 || score > 1000 {
		return badRequest("score must be between -1000 and 1000")
	}
	return nil
}
