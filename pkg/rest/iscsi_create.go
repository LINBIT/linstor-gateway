package rest

import (
	"net/http"

	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
)

// ISCSICreate creates a highly-available iSCSI target via the REST-API
func ISCSICreate(w http.ResponseWriter, r *http.Request) {
	var iscsiCfg iscsi.ISCSI
	if err := unmarshalBody(w, r, &iscsiCfg); err != nil {
		return
	}
	maybeSetLinstorController(&iscsiCfg)

	if err := iscsi.CheckIQN(iscsiCfg.Target.IQN); err != nil {
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
