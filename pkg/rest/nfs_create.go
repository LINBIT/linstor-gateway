package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

// NFSCreate creates a highly-available NFS export via the REST-API
func (s *server) NFSCreate() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var rsc nfs.ResourceConfig
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to parse request body: %v", err)
			return
		}

		result, err := nfs.Create(request.Context(), &rsc)
		if err != nil {
			_, _ = Errorf(http.StatusBadRequest, writer, "failed to create nfs resource: %v", err)
			return
		}

		writer.Header().Add("Location", fmt.Sprintf("./nfs/%s", result.Name))
		writer.WriteHeader(http.StatusCreated)
		encoder := json.NewEncoder(writer)

		err = encoder.Encode(result)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
