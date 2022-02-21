package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// HealthCheck used for checking HTTP APIs are available or not
func (s *server) APIStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "ok",
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)

		err := enc.Encode(status)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
