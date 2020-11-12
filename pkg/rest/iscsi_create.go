package rest

import (
	"net/http"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	log "github.com/sirupsen/logrus"
)

// ISCSICreate creates a highly-available iSCSI target via the REST-API
func (s *server) ISCSICreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		defer s.Unlock()

		var iscsiCfg iscsi.ISCSI
		if err := unmarshalBody(w, r, &iscsiCfg); err != nil {
			return
		}
		maybeSetLinstorController(&iscsiCfg)

		log.Debugf("Unmarshalled config: %+v", iscsiCfg)

		if err := targetutil.CheckIQN(iscsiCfg.Target.IQN); err != nil {
			_, _ = Errorf(http.StatusBadRequest, w, "Could not validate IQN: %v", err)
			return
		}
		if err := iscsiCfg.CreateResource(); err != nil {
			_, _ = Errorf(http.StatusInternalServerError, w, "Could not create resource: %v", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		// json.NewEncoder(w).Encode(xxx)
	}
}
