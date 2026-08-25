package logic

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"class-management-system/backend/internal/httperr"
)

func TestValidateStudentFields(t *testing.T) {
	if err := validateStudentFields("2026001", "张三", "男", "13800000000", "班长"); err != nil {
		t.Fatalf("validateStudentFields() rejected valid fields: %v", err)
	}

	tests := []struct {
		name      string
		studentNo string
		student   string
		gender    string
		phone     string
		position  string
	}{
		{name: "blank student number", studentNo: " ", student: "张三"},
		{name: "blank name", studentNo: "2026001", student: " "},
		{name: "long student number", studentNo: strings.Repeat("1", 65), student: "张三"},
		{name: "long name", studentNo: "2026001", student: strings.Repeat("名", 65)},
		{name: "long gender", studentNo: "2026001", student: "张三", gender: strings.Repeat("性", 17)},
		{name: "long phone", studentNo: "2026001", student: "张三", phone: strings.Repeat("1", 33)},
		{name: "long position", studentNo: "2026001", student: "张三", position: strings.Repeat("职", 65)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStudentFields(tt.studentNo, tt.student, tt.gender, tt.phone, tt.position)
			var httpErr *httperr.Error
			if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
				t.Fatalf("validateStudentFields() error = %v, want HTTP 400", err)
			}
		})
	}
}

func TestValidateScoreValue(t *testing.T) {
	for _, score := range []int64{-1000, 0, 1000} {
		if err := validateScoreValue(score); err != nil {
			t.Fatalf("validateScoreValue(%d) error = %v", score, err)
		}
	}
	for _, score := range []int64{-1001, 1001} {
		if err := validateScoreValue(score); err == nil {
			t.Fatalf("validateScoreValue(%d) accepted out-of-range score", score)
		}
	}
}
