package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func (s *server) ISCSIAddVolume() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		iqn, err := iscsi.NewIqn(mux.Vars(request)["iqn"])
		if err != nil {
			MustError(http.StatusBadRequest, writer, "malformed iqn: %v", err)
			return
		}

		lun, err := strconv.Atoi(mux.Vars(request)["lun"])
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "malformed LUN: %v", err)
			return
		}

		var vCfg common.VolumeConfig
		decoder := json.NewDecoder(request.Body)
		err = decoder.Decode(&vCfg)
		if err != nil {
			MustError(http.StatusBadRequest, writer, "failed to parse request body: %v", err)
			return
		}

		if vCfg.Number != 0 && vCfg.Number != lun {
			MustError(http.StatusBadRequest, writer, "expected volume number to be %d, but request body has %d", lun, vCfg.Number)
			return
		}

		// Fill in default
		vCfg.Number = lun

		if lun < 1 {
			MustError(http.StatusBadRequest, writer, "volume number must be positive, is %d", lun)
			return
		}

		cfg, err := s.iscsi.AddVolume(ctx, iqn, &vCfg)
		if err != nil {
			MustError(http.StatusInternalServerError, writer, "failed to add volume to resource: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, writer, "no resource found for iqn %s", iqn)
			return
		}

		volCfg := cfg.VolumeConfig(lun)

		writer.WriteHeader(http.StatusOK)
		err = json.NewEncoder(writer).Encode(volCfg)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
