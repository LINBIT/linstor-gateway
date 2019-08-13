package rest

import (
	"net/http"

	"github.com/LINBIT/linstor-remote-storage/iscsi"
)

func ISCSIDelete(w http.ResponseWriter, r *http.Request) {
	var iscsiCfg iscsi.ISCSI
	if err := unmarshalBody(w, r, &iscsiCfg); err != nil {
		return
	}

	if err := iscsiCfg.DeleteResource(); err != nil {
		Errorf(http.StatusInternalServerError, w, "Could not delete resource: %v", err)
		return
	}

	// json.NewEncoder(w).Encode(xxx)
}
