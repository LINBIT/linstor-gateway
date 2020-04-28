package rest

import "net/http"

func (s *server) ISCSIStart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		defer s.Unlock()

		iscsiCfg, ok := parseIQNAndLun(w, r)
		if !ok {
			return
		}

		if err := iscsiCfg.StartResource(); err != nil {
			_, _ = Errorf(http.StatusInternalServerError, w, "Could not start resource: %v", err)
			return
		}
	}
}
