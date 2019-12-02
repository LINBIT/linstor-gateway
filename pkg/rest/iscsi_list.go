package rest

import (
	"encoding/json"
	"net/http"

	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
)

func (s *server) ISCSIList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		defer s.Unlock()

		targets, err := iscsi.ListResources()
		if err != nil {
			_, _ = Errorf(http.StatusInternalServerError, w, "Could not list targets: %w", err)
			return
		}

		json.NewEncoder(w).Encode(targets)
	}
}
