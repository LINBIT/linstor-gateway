package rest

import (
	"net/http"

	"github.com/LINBIT/linstor-iscsi/pkg/nfs"
)

func (srv *server) NFSCreate() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		srv.Lock()
		defer srv.Unlock()

		var nfsRsc nfs.NfsResource
		if err := unmarshalBody(response, request, &nfsRsc); err != nil {
			_, _ = Errorf(http.StatusBadRequest, response, "Cannot unmarshal JSON data: %v", err)
			return
		}
		maybeSetLinstorController(&nfsRsc)

		if err := nfsRsc.CreateResource(); err != nil {
			_, _ = Errorf(http.StatusInternalServerError, response, "Could not create resource: %v", err)
			return
		}

		response.WriteHeader(http.StatusCreated)
	}
}
