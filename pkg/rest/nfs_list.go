package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

func (s *server) NFSList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets, err := nfs.List(r.Context())
		if err != nil {
			MustError(http.StatusInternalServerError, w, "Could not list exports: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)

		err = enc.Encode(targets)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
