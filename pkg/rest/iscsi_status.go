package rest

import (
	"encoding/json"
	"net/http"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
)

type state struct {
	Target  crmcontrol.ResourceRunState     `json:"pacemaker,omitempty"`
	Linstor linstorcontrol.ResourceRunState `json:"linstor,omitempty"`
}

func ISCSIStatus(w http.ResponseWriter, r *http.Request) {
	restMutex.Lock()
	defer restMutex.Unlock()

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
