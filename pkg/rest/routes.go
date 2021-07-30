package rest

import "net/http"

func (s *server) routes() {
	apiv2 := s.router.PathPrefix("/api/v2").Subrouter()
	apiv2.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			handler.ServeHTTP(w, r)
		})
	})

	/*
	iscsiv1 := apiv1.PathPrefix("/iscsi").Subrouter()
	iscsiv1.HandleFunc("", s.ISCSICreate()).Methods("POST")
	iscsiv1.HandleFunc("", s.ISCSIList()).Methods("GET")
	iscsiv1.HandleFunc("/{iqn}/{lun}", s.ISCSIDelete()).Methods("DELETE")
	iscsiv1.HandleFunc("/{iqn}/{lun}", s.ISCSIStatus()).Methods("GET")
	iscsiv1.HandleFunc("/{iqn}/{lun}/start", s.ISCSIStart()).Methods("POST")
	iscsiv1.HandleFunc("/{iqn}/{lun}/stop", s.ISCSIStop()).Methods("POST")

	nfsv1 := apiv1.PathPrefix("/nfs").Subrouter()
	nfsv1.HandleFunc("", s.NFSCreate()).Methods("POST")
	nfsv1.HandleFunc("", s.NFSList()).Methods("GET")
	nfsv1.HandleFunc("/{resource}", s.NFSDelete()).Methods("DELETE")
	nfsv1.HandleFunc("/{resource}", s.NFSStatus()).Methods("GET")
	*/

	nvmeofv1 := apiv2.PathPrefix("/nvme-of").Subrouter()
	nvmeofv1.HandleFunc("", s.NVMeoFList()).Methods("GET")
	nvmeofv1.HandleFunc("", s.NVMeoFCreate()).Methods("POST")
	nvmeofv1.HandleFunc("/{nqn}", s.NVMeoFGet(true)).Methods("GET")
	nvmeofv1.HandleFunc("/{nqn}", s.NVMeoFDelete(true)).Methods("DELETE")
	nvmeofv1.HandleFunc("/{nqn}/start", s.NVMeoFStart()).Methods("POST")
	nvmeofv1.HandleFunc("/{nqn}/stop", s.NVMeoFStop()).Methods("POST")
	nvmeofv1.HandleFunc("/{nqn}/{nsid}", s.NVMeoFGet(false)).Methods("GET")
	nvmeofv1.HandleFunc("/{nqn}/{nsid}", s.NVMeoFAddVolume()).Methods("PUT")
	nvmeofv1.HandleFunc("/{nqn}/{nsid}", s.NVMeoFDelete(false)).Methods("DELETE")

}
