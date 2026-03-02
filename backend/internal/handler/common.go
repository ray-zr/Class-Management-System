package handler

import (
	"net/http"
	"strconv"

	"class-management-system/backend/internal/httperr"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

func pathInt64(r *http.Request, key string) (int64, error) {
	val := r.PathValue(key)
	if val == "" {
		val = pathvar.Vars(r)[key]
	}
	if val == "" {
		return 0, &httperr.Error{Code: http.StatusBadRequest, Msg: "missing path param"}
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, &httperr.Error{Code: http.StatusBadRequest, Msg: "invalid path param"}
	}
	return id, nil
}
