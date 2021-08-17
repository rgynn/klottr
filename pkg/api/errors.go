package api

import (
	"encoding/json"
	"net/http"

	"github.com/rgynn/ptrconv"
	"github.com/sirupsen/logrus"
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

	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		logger = logrus.New().WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.Query().Encode(),
		})
	}

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
		if _, err := w.Write([]byte(err.Error())); err != nil {
			logger.Debug(err)
		}
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			logger.Debug(err)
		}
		return
	}
}
