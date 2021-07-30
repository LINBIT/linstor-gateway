package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFCreate() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		var rsc nvmeof.ResourceConfig
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to parse request body: %v", err)
			return
		}

		result, err := nvmeof.Create(request.Context(), &rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to create nvmeof resource: %v", err)
			return
		}

		writer.Header().Add("Location", fmt.Sprintf("./nvme-of/%s", result.NQN))
		writer.WriteHeader(http.StatusCreated)
		encoder := json.NewEncoder(writer)

		err = encoder.Encode(result)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
