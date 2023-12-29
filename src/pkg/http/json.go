package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type JSONErrorResponse struct {
	Details string `json:"details"`
}

func RespondWithJSONError(l *zap.Logger, w http.ResponseWriter, err error) {
	httpErr := DetermineHTTPError(err)
	w.WriteHeader(httpErr.StatusCode)
	RespondWithJSON(l, w, JSONErrorResponse{Details: httpErr.Details})
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
