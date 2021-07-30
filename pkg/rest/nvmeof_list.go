package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFList() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		cfgs, err := nvmeof.List(ctx)
		if err != nil {
			_, err := Errorf(http.StatusInternalServerError, writer, "nvmeof list failed: %v", err)
			if err != nil {
				log.WithError(err).Warn("failed to write error response")
			}
			return
		}

		writer.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(writer)

		err = enc.Encode(cfgs)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
