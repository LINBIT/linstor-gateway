package rest

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func (s *server) ISCSIList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets, err := iscsi.List(r.Context())
		if err != nil {
			MustError(http.StatusInternalServerError, w, "Could not list targets: %v", err)
			return
		}

		for i := range targets {
			targets[i].Username = ""
			targets[i].Password = ""
		}

		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)

		err = enc.Encode(targets)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
