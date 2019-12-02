package rest

func (s *server) routes() {
	s.router.HandleFunc("/api/v1/iscsi", s.ISCSICreate()).Methods("POST")
	s.router.HandleFunc("/api/v1/iscsi", s.ISCSIList()).Methods("GET")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}", s.ISCSIDelete()).Methods("DELETE")
	s.router.HandleFunc("/api/v1/iscsi/{iqn}/{lun}", s.ISCSIStatus()).Methods("GET")
}
