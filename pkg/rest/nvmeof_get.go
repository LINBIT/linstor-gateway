package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFGet(all bool) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		nqn, err := nvmeof.NewNqn(mux.Vars(request)["nqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed nqn: %v", err)
			return
		}

		cfg, err := s.nvmeof.Get(ctx, nqn)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to fetch resource status: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, writer, "no resource found for nqn %s", nqn)
			return
		}

		if all {
			writer.WriteHeader(http.StatusOK)
			err = json.NewEncoder(writer).Encode(cfg)
			if err != nil {
				log.WithError(err).Warn("failed to write response")
			}
		} else {
			nsid, err := strconv.Atoi(mux.Vars(request)["nsid"])
			if err != nil {
				MustError(http.StatusInternalServerError, writer, "wrong namespace id format: %v", err)
				return
			}

			volCfg := cfg.VolumeConfig(nsid)
			if volCfg == nil {
				MustError(http.StatusNotFound, writer, "no volume found for nqn %s, nsid %d", nqn, nsid)
				return
			}

			writer.WriteHeader(http.StatusOK)
			err = json.NewEncoder(writer).Encode(volCfg)
			if err != nil {
				log.WithError(err).Warn("failed to write response")
			}
		}
	}
}
