package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (s *server) NFSStop() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resource := mux.Vars(request)["resource"]

		cfg, err := s.nfs.Stop(request.Context(), resource)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to stop export: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, writer, "no resource found")
			return
		}

		writer.Header().Add("Location", "./")
		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(cfg)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
