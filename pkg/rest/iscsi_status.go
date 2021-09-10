package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func (s *server) ISCSIGet(all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		iqn, err := iscsi.NewIqn(mux.Vars(r)["iqn"])
		if err != nil {
			MustError(http.StatusBadRequest, w, "malformed iqn: %v", err)
			return
		}

		cfg, err := s.iscsi.Get(r.Context(), iqn)
		if err != nil {
			MustError(http.StatusInternalServerError, w, "failed to fetch resource status: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, w, "no resource found for iqn %s", iqn)
			return
		}

		if all {
			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)

			err = enc.Encode(cfg)
			if err != nil {
				log.WithError(err).Warn("failed to write response")
			}
		} else {
			lun, err := strconv.Atoi(mux.Vars(r)["lun"])
			if err != nil {
				MustError(http.StatusBadRequest, w, "invalid LUN: %v", err)
				return
			}

			vol := cfg.VolumeConfig(lun)
			if vol == nil {
				MustError(http.StatusNotFound, w, "no volume found for iqn %s, lun %d", iqn, lun)
				return
			}

			w.WriteHeader(http.StatusOK)
			enc := json.NewEncoder(w)

			err = enc.Encode(vol)
			if err != nil {
				log.WithError(err).Warn("failed to write response")
			}
		}
	}
}
