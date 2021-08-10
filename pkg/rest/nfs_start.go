package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

func (s *server) NFSStart() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resource := mux.Vars(request)["resource"]

		cfg, err := nfs.Start(request.Context(), resource)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to start export: %v", err)
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
