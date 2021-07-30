package rest

import (
	"encoding/json"
	"net/http"

	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
)

func (s *server) ISCSIStatus() http.HandlerFunc {
	type state struct {
		Target  crmcontrol.ResourceRunState     `json:"pacemaker,omitempty"`
		Linstor linstorcontrol.ResourceRunState `json:"linstor,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		defer s.Unlock()

		iscsiCfg, ok := parseIQNAndLun(w, r)
		if !ok {
			return
		}

		tstate, err := iscsiCfg.ProbeResource()
		if err != nil {
			_, _ = Errorf(http.StatusInternalServerError, w, "Could not retrieve status: %v", err)
			return
		}

		lrstate, err := iscsiCfg.Linstor.AggregateResourceState()
		if err != nil {
			lrstate = linstorcontrol.Unknown
		}

		s := state{
			Target:  tstate,
			Linstor: linstorcontrol.ResourceRunState{ResourceState: lrstate},
		}

		json.NewEncoder(w).Encode(s)
	}
}
