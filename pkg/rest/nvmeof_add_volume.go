package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func (s *server) NVMeoFAddVolume() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		nqn, err := nvmeof.NewNqn(mux.Vars(request)["nqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed nqn: %v", err)
			return
		}

		nsid, err := strconv.Atoi(mux.Vars(request)["nsid"])
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "wrong namespace id format: %v", err)
			return
		}

		var vCfg common.VolumeConfig
		decoder := json.NewDecoder(request.Body)
		err = decoder.Decode(&vCfg)
		if err != nil {
			MustError(http.StatusBadRequest, writer, "failed to parse request body: %v", err)
			return
		}

		if vCfg.Number != 0 && vCfg.Number != nsid {
			MustError(http.StatusBadRequest, writer, "expected volume number to be %d, but request body has %d", nsid, vCfg.Number)
			return
		}

		// Fill in default
		vCfg.Number = nsid

		if nsid < 1 {
			MustError(http.StatusBadRequest, writer, "volume number must be positive, is %d", nsid)
			return
		}

		cfg, err := s.nvmeof.AddVolume(ctx, nqn, &vCfg)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to add volume to resource: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, writer, "no resource found for nqn %s", nqn)
			return
		}

		volCfg := cfg.VolumeConfig(nsid)

		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(volCfg)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
