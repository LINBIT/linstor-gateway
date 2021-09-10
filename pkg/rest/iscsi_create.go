package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

// ISCSICreate creates a highly-available iSCSI target via the REST-API
func (s *server) ISCSICreate() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var rsc iscsi.ResourceConfig
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to parse request body: %v", err)
			return
		}

		result, err := s.iscsi.Create(request.Context(), &rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to create iscsi resource: %v", err)
			return
		}

		writer.Header().Add("Location", fmt.Sprintf("./iscsi/%s", result.IQN))
		writer.WriteHeader(http.StatusCreated)
		encoder := json.NewEncoder(writer)

		err = encoder.Encode(result)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
