package rest

import (
	"encoding/json"
	"net/http"

	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
)

type NfsState struct {
	CrmState     crmcontrol.NfsRunState       `json:"crm_state,omitempty"`
	LinstorState linstorcontrol.ResourceState `json:"linstor_state,omitempty"`
}

func (srv *server) NFSStatus() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		srv.Lock()
		defer srv.Unlock()

		nfsRsc, ok := parseNFSResource(response, request)
		if !ok {
			// Error reported by function call
			return
		}

		crmState, err := nfsRsc.ProbeResource()
		if err != nil {
			_, _ = Errorf(http.StatusInternalServerError, response, "Failed to determine resource status: %v", err)
			return
		}

		linstorState, err := nfsRsc.Linstor.AggregateResourceState()
		if err != nil {
			linstorState = linstorcontrol.Unknown
		}

		state := NfsState{crmState, linstorState}

		json.NewEncoder(response).Encode(state)
	}
}
