package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// NFSDelete deletes a highly-available NFS export via the REST-API
func (s *server) NFSDelete(all bool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		resource := mux.Vars(request)["resource"]

		if all {
			err := s.nfs.Delete(ctx, resource)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "delete failed: %v", err)
				return
			}
		} else {
			id, err := strconv.Atoi(mux.Vars(request)["id"])
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "invalid volume: %v", err)
				return
			}

			oldCfg, err := s.nfs.DeleteVolume(ctx, resource, id)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "error deleting volume: %v", err)
				return
			}

			if oldCfg == nil {
				MustError(http.StatusNotFound, writer, "no resource found")
				return
			}
		}

		writer.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(writer)

		err := enc.Encode(struct{}{})
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
