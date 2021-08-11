package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse for api
type ErrorResponse struct {
	Path    string `json:"path"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {

	body, err := json.Marshal(ErrorResponse{
		Path:    r.URL.RawPath,
		Code:    status,
		Message: err.Error(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
