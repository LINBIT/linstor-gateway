package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

// ISCSIDelete deletes a highly-available iSCSI target via the REST-API
func (s *server) ISCSIDelete(all bool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		iqn, err := iscsi.NewIqn(mux.Vars(request)["iqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed iqn: %v", err)
			return
		}

		if all {
			err = s.iscsi.Delete(ctx, iqn)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "delete failed: %v", err)
				return
			}
		} else {
			lun, err := strconv.Atoi(mux.Vars(request)["lun"])
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "invalid LUN: %v", err)
				return
			}

			oldCfg, err := s.iscsi.DeleteVolume(ctx, iqn, lun)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "error deleting volume: %v", err)
				return
			}

			if oldCfg == nil {
				MustError(http.StatusNotFound, writer, "no resource found for iqn %s", iqn)
				return
			}
		}

		writer.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(writer)

		err = enc.Encode(struct{}{})
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
