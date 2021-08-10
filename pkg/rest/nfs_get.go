package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

func (s *server) NFSGet(all bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resource := mux.Vars(r)["resource"]

		cfg, err := nfs.Get(r.Context(), resource)
		if err != nil {
			MustError(http.StatusInternalServerError, w, "failed to fetch resource status: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, w, "no resource found")
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
			id, err := strconv.Atoi(mux.Vars(r)["id"])
			if err != nil {
				MustError(http.StatusBadRequest, w, "invalid LUN: %v", err)
				return
			}

			vol := cfg.VolumeConfig(id)
			if vol == nil {
				MustError(http.StatusNotFound, w, "no volume found for resource %s, export %d", resource, id)
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
