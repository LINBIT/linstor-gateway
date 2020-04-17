package rest

import (
	"net/http"
	"github.com/LINBIT/linstor-iscsi/pkg/nfs"
)

func (srv *server) NFSDelete() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		srv.Lock()
		defer srv.Unlock()

		var nfsRsc nfs.NFSResource
		// TODO: Parse resource name from JSON
		//       See parseIQNAndLUN

		if err := nfsRsc.DeleteResource(); err != nil {
			_, _ = Errorf(http.StatusInternalServerError, response, "Could not delete resource: %v", err)
			return
		}
	}
}
