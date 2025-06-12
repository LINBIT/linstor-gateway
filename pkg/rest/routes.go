package rest

import (
	"net/http"

	"github.com/LINBIT/linstor-gateway/pkg/version"
)

// serverNameMiddleware adds a "Server" header to the response, identifying the linstor-gateway server.
func serverNameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "linstor-gateway/"+version.Version)
		next.ServeHTTP(w, r)
	})
}

func (s *server) routes() {
	s.router.Use(serverNameMiddleware)

	apiv2 := s.router.PathPrefix("/api/v2").Subrouter()
	apiv2.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			handler.ServeHTTP(w, r)
		})
	})

	apiv2.HandleFunc("/status", s.APIStatus()).Methods("GET")

	iscsiv2 := apiv2.PathPrefix("/iscsi").Subrouter()
	iscsiv2.HandleFunc("", s.ISCSIList()).Methods("GET")
	iscsiv2.HandleFunc("", s.ISCSICreate()).Methods("POST")
	iscsiv2.HandleFunc("/{iqn}", s.ISCSIGet(true)).Methods("GET")
	iscsiv2.HandleFunc("/{iqn}", s.ISCSIDelete(true)).Methods("DELETE")
	iscsiv2.HandleFunc("/{iqn}/start", s.ISCSIStart()).Methods("POST")
	iscsiv2.HandleFunc("/{iqn}/stop", s.ISCSIStop()).Methods("POST")
	iscsiv2.HandleFunc("/{iqn}/{lun}", s.ISCSIGet(false)).Methods("GET")
	iscsiv2.HandleFunc("/{iqn}/{lun}", s.ISCSIAddVolume()).Methods("PUT")
	iscsiv2.HandleFunc("/{iqn}/{lun}", s.ISCSIDelete(false)).Methods("DELETE")

	nfsv2 := apiv2.PathPrefix("/nfs").Subrouter()
	nfsv2.HandleFunc("", s.NFSList()).Methods("GET")
	nfsv2.HandleFunc("", s.NFSCreate()).Methods("POST")
	nfsv2.HandleFunc("/{resource}", s.NFSGet(true)).Methods("GET")
	nfsv2.HandleFunc("/{resource}", s.NFSDelete(true)).Methods("DELETE")
	nfsv2.HandleFunc("/{resource}/start", s.NFSStart()).Methods("POST")
	nfsv2.HandleFunc("/{resource}/stop", s.NFSStop()).Methods("POST")
	nfsv2.HandleFunc("/{resource}/{id}", s.NFSGet(false)).Methods("GET")
	// No add volume: LINSTOR refuses to create a filesystem on volume that are added after the resource is deployed.
	nfsv2.HandleFunc("/{resource}/{id}", s.NFSDelete(false)).Methods("DELETE")

	nvmeofv2 := apiv2.PathPrefix("/nvme-of").Subrouter()
	nvmeofv2.HandleFunc("", s.NVMeoFList()).Methods("GET")
	nvmeofv2.HandleFunc("", s.NVMeoFCreate()).Methods("POST")
	nvmeofv2.HandleFunc("/{nqn}", s.NVMeoFGet(true)).Methods("GET")
	nvmeofv2.HandleFunc("/{nqn}", s.NVMeoFDelete(true)).Methods("DELETE")
	nvmeofv2.HandleFunc("/{nqn}/start", s.NVMeoFStart()).Methods("POST")
	nvmeofv2.HandleFunc("/{nqn}/stop", s.NVMeoFStop()).Methods("POST")
	nvmeofv2.HandleFunc("/{nqn}/{nsid}", s.NVMeoFGet(false)).Methods("GET")
	nvmeofv2.HandleFunc("/{nqn}/{nsid}", s.NVMeoFAddVolume()).Methods("PUT")
	nvmeofv2.HandleFunc("/{nqn}/{nsid}", s.NVMeoFDelete(false)).Methods("DELETE")

	// gorilla/mux usually does not apply middlewares to the NotFoundHandler. To apply the serverNameMiddleware,
	// overwrite the NotFoundHandler with a new route that has the middleware applied.
	s.router.NotFoundHandler = s.router.NewRoute().HandlerFunc(http.NotFound).GetHandler()
}
