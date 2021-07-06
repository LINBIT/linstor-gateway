package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func (s *server) ISCSIStop() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		iqn, err := iscsi.NewIqn(mux.Vars(r)["iqn"])
		if err != nil {
			MustError(http.StatusBadRequest, w, "malformed iqn: %v", err)
			return
		}

		cfg, err := iscsi.Stop(ctx, iqn)
		if err != nil {
			MustError(http.StatusInternalServerError, w, "failed to stop resource: %v", err)
			return
		}

		if cfg == nil {
			MustError(http.StatusNotFound, w, "no resource found for iqn %s", iqn)
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
