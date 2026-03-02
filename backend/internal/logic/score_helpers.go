package logic

import (
	"net/http"

	"class-management-system/backend/internal/httperr"
)

func badRequest(msg string) error {
	return &httperr.Error{Code: http.StatusBadRequest, Msg: msg}
}
