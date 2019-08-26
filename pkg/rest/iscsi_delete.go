package rest

import (
	"net/http"
	"strconv"

	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
	"github.com/gorilla/mux"
)

// ISCSIDelete deletes a highly-available iSCSI target via the REST-API
func ISCSIDelete(w http.ResponseWriter, r *http.Request) {
	restMutex.Lock()
	defer restMutex.Unlock()

	tgt, ok := mux.Vars(r)["target"]
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
		IQN:  "iqn.1981-09.at.rck:" + tgt,
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
	if err := iscsiCfg.DeleteResource(); err != nil {
		_, _ = Errorf(http.StatusInternalServerError, w, "Could not delete resource: %v", err)
		return
	}

	// json.NewEncoder(w).Encode(xxx)
}
