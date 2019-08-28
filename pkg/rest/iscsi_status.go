package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
	"github.com/gorilla/mux"
)

type state struct {
	Target  crmcontrol.ResourceRunState     `json:"pacemaker,omitempty"`
	Linstor linstorcontrol.ResourceRunState `json:"linstor,omitempty"`
}

func ISCSIStatus(w http.ResponseWriter, r *http.Request) {
	restMutex.Lock()
	defer restMutex.Unlock()

	iqn, ok := mux.Vars(r)["iqn"]
	if !ok {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not find 'target' in your request")
		return
	}

	l, ok := mux.Vars(r)["lun"]
	if !ok {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not find 'lun' in your request")
		return
	}
	lid, err := strconv.Atoi(l)
	if err != nil {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not convert LUN to number: %v", err)
		return
	}

	lun := targetutil.LUN{ID: uint8(lid)}
	targetConfig := targetutil.TargetConfig{
		IQN:  iqn,
		LUNs: []*targetutil.LUN{&lun},
	}
	target, err := targetutil.NewTarget(targetConfig)
	if err != nil {
		_, _ = Errorf(http.StatusInternalServerError, w, "Could not create target from target config: %v", err)
		return
	}

	iscsiCfg := iscsi.ISCSI{
		Target:  target,
		Linstor: linstorcontrol.Linstor{},
	}

	maybeSetLinstorController(&iscsiCfg)

	if err := targetutil.CheckIQN(iscsiCfg.Target.IQN); err != nil {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not validate IQN: %v", err)
		return
	}

	tstate, err := iscsiCfg.ProbeResource()
	if err != nil {
		_, _ = Errorf(http.StatusInternalServerError, w, "Could not retrieve status: %v", err)
		return
	}

	targetName, _ := targetutil.ExtractTargetName(target.IQN)
	iscsiCfg.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, uint8(lid))
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
