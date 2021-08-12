package api

import (
	"encoding/json"
	"net/http"

	"github.com/rgynn/ptrconv"
)

// ErrorResponse for api
type ErrorResponse struct {
	RequestID *string `json:"reqid,omitempty"`
	Method    string  `json:"method"`
	Path      string  `json:"path"`
	Query     string  `json:"query,omitempty"`
	Code      int     `json:"code"`
	Message   string  `json:"message"`
}

func NewErrorResponse(w http.ResponseWriter, r *http.Request, status int, inputerr error) {

	reqid, err := RequestIDFromContext(r.Context())
	if err != nil {
		reqid = ptrconv.StringPtr("none")
	}

	body, err := json.Marshal(ErrorResponse{
		RequestID: reqid,
		Method:    r.Method,
		Path:      r.URL.Path,
		Query:     r.URL.Query().Encode(),
		Code:      status,
		Message:   inputerr.Error(),
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
