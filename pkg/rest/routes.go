package rest

func (s *server) routes() {
	s.router.HandleFunc("/api/v1/iscsi", s.ISCSICreate()).Methods("POST")
	s.router.HandleFunc("/api/v1/iscsi", s.ISCSIList()).Methods("GET")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}", s.ISCSIDelete()).Methods("DELETE")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}", s.ISCSIStatus()).Methods("GET")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}/start", s.ISCSIStart()).Methods("POST")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}/stop", s.ISCSIStop()).Methods("POST")
	s.router.HandleFunc("/api/v1/nfs", s.NFSCreate()).Methods("POST")
	s.router.HandleFunc("/api/v1/nfs", s.NFSList()).Methods("GET")
	s.router.HandleFunc("/api/v1/nfs/{resource}", s.NFSDelete()).Methods("DELETE")
	s.router.HandleFunc("/api/v1/nfs/{resource}", s.NFSStatus()).Methods("GET")
}
