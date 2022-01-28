package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFDelete(all bool) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		nqn, err := nvmeof.NewNqn(mux.Vars(request)["nqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed nqn: %v", err)
			return
		}

		if all {
			deployed, err := s.nvmeof.Get(ctx, nqn)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "failed to query target: %v", err)
			}
			if deployed == nil {
				MustError(http.StatusNotFound, writer, "no resource found for nqn %s", nqn)
			}
			err = s.nvmeof.Delete(ctx, nqn)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "nvmeof delete failed: %v", err)
				return
			}
		} else {
			nsid, err := strconv.Atoi(mux.Vars(request)["nsid"])
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "wrong namespace id format: %v", err)
				return
			}

			oldCfg, err := s.nvmeof.DeleteVolume(ctx, nqn, nsid)
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "error deleting volume: %v", err)
				return
			}

			if oldCfg == nil {
				MustError(http.StatusNotFound, writer, "no resource found for nqn %s", nqn)
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
