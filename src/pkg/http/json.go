package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type JSONErrorResponse struct {
	Details string
}

func RespondWithJSONError(l *zap.Logger, w http.ResponseWriter, err error) {
	var (
		errorCode = 500
		details   = err.Error()
	)

	var httpErr Err
	if errors.As(err, &httpErr) {
		errorCode = httpErr.code
		if c := httpErr.cause; c != nil {
			details = c.Error()
		}
	}

	w.WriteHeader(errorCode)
	RespondWithJSON(l, w, JSONErrorResponse{Details: details})
}

// RespondWithJSON serializes the val into JSON and writes it into the response writer.
// Expects that if any status code was required to set, it is already set for the writer.
func RespondWithJSON(l *zap.Logger, w http.ResponseWriter, val interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add(HeaderContentType, ContentTypeJSON)

	var err error
	if str, ok := val.(string); ok {
		_, err = w.Write([]byte(str))
	} else {
		err = json.NewEncoder(w).Encode(val)
	}

	if err != nil {
		l.Error("failed to write response into request", zap.Error(err))
	}
}

func DefaultOKResponse() string {
	return `{"status": "ok"}`
}
