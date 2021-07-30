package rest

import (
	"net/http"
)

// ISCSIDelete deletes a highly-available iSCSI target via the REST-API
func (s *server) ISCSIDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		defer s.Unlock()

		iscsiCfg, ok := parseIQNAndLun(w, r)
		if !ok {
			return
		}

		if err := iscsiCfg.DeleteResource(); err != nil {
			_, _ = Errorf(http.StatusInternalServerError, w, "Could not delete resource: %v", err)
			return
		}
		// json.NewEncoder(w).Encode(xxx)
	}
}
