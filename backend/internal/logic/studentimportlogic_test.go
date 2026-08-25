package logic

import (
	"strings"
	"testing"
)

func TestStudentsFromRows(t *testing.T) {
	rows := [][]string{
		{" 2026001 ", " 张三 ", "男", "13800000000", "班长"},
		{"2026001", "李四"},
		{"2026002", ""},
	}
	students, rowErrors := studentsFromRows(rows)
	if len(students) != 1 {
		t.Fatalf("studentsFromRows() returned %d students, want 1", len(students))
	}
	if students[0].StudentNo != "2026001" || students[0].Name != "张三" {
		t.Fatalf("studentsFromRows() did not trim fields: %+v", students[0])
	}
	if len(rowErrors) != 2 {
		t.Fatalf("studentsFromRows() returned %d errors, want 2: %v", len(rowErrors), rowErrors)
	}
	if !strings.Contains(rowErrors[0], "row 3") || !strings.Contains(rowErrors[0], "duplicate studentNo") {
		t.Fatalf("duplicate error = %q", rowErrors[0])
	}
	if !strings.Contains(rowErrors[1], "row 4") {
		t.Fatalf("validation error = %q", rowErrors[1])
	}
}

func TestFormatImportErrorsCapsDetails(t *testing.T) {
	errors := make([]string, 25)
	for i := range errors {
		errors[i] = "invalid row"
	}
	message := formatImportErrors(errors)
	if strings.Count(message, "invalid row") != 20 {
		t.Fatalf("formatImportErrors() included the wrong number of details: %q", message)
	}
	if !strings.HasSuffix(message, "and 5 more error(s)") {
		t.Fatalf("formatImportErrors() summary = %q", message)
	}
}
