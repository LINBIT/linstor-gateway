package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFStop() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		nqn, err := nvmeof.NewNqn(mux.Vars(request)["nqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed nqn: %v", err)
			return
		}

		resourceTimeout, err := parseResourceTimeout(request)
		if err != nil {
			MustError(http.StatusBadRequest, writer, "invalid resource_timeout: %v", err)
			return
		}

		cfg, err := s.nvmeof.Stop(ctx, nqn, resourceTimeout)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to stop resource: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, writer, "no resource found for nqn %s", nqn)
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
