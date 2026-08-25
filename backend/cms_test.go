package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"class-management-system/backend/internal/httperr"

	"github.com/go-sql-driver/mysql"
)

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{name: "expected HTTP error", err: &httperr.Error{Code: http.StatusUnauthorized, Msg: "unauthorized"}, wantStatus: http.StatusUnauthorized, wantMessage: "unauthorized"},
		{name: "malformed JSON", err: &json.SyntaxError{}, wantStatus: http.StatusBadRequest, wantMessage: "invalid request"},
		{name: "duplicate record", err: &mysql.MySQLError{Number: 1062}, wantStatus: http.StatusConflict, wantMessage: "record already exists"},
		{name: "internal error is redacted", err: errors.New("database host and details"), wantStatus: http.StatusInternalServerError, wantMessage: "internal server error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := errorResponse(tt.err)
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", status, tt.wantStatus)
			}
			payload, ok := body.(map[string]any)
			if !ok {
				t.Fatalf("body type = %T, want map[string]any", body)
			}
			if payload["message"] != tt.wantMessage {
				t.Fatalf("message = %v, want %q", payload["message"], tt.wantMessage)
			}
		})
	}
}
