package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func (s *server) ISCSIStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		iqn, err := iscsi.NewIqn(mux.Vars(r)["iqn"])
		if err != nil {
			MustError(http.StatusBadRequest, w, "invalid iqn: %v", err)
			return
		}

		cfg, err := s.iscsi.Start(r.Context(), iqn)
		if err != nil {
			MustError(http.StatusInternalServerError, w, "failed to start target: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, w, "no resource with iqn %s found", iqn)
			return
		}

		w.Header().Add("Location", "./")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(cfg)
		if err != nil {
			log.WithError(err).Warn("failed to write response")
		}
	}
}
