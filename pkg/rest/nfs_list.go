package rest

import (
	"encoding/json"
	"net/http"

	"github.com/LINBIT/linstor-gateway/pkg/nfs"
)

func (srv *server) NFSList() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		srv.Lock()
		defer srv.Unlock()

		nfsRscList, err := nfs.ListResources()
		if err != nil {
			_, _ = Errorf(http.StatusInternalServerError, response, "Failed to list NFS resources: %v", err)
			return
		}

		json.NewEncoder(response).Encode(nfsRscList)
	}
}
