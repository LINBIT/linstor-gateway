// Package rest provides the REST API to create highly-available iSCSI targets.
package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type server struct {
	router *mux.Router
	iscsi  *iscsi.ISCSI
	nfs    *nfs.NFS
	nvmeof *nvmeof.NVMeoF
	sync.Mutex
}

// Error is the type that is returned in case of an error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

// Errorf takes a StatusCode, a ResponseWriter and a format string.
// It sets up the REST response and writes it to the ResponseWriter
// It also sets the according error code.
func Errorf(code int, w http.ResponseWriter, format string, a ...interface{}) (n int, err error) {
	e := Error{
		Code:    http.StatusText(code),
		Message: fmt.Sprintf(format, a...),
	}

	b, err := json.Marshal(&e)
	if err != nil {
		return 0, err
	}

	w.WriteHeader(code)
	return fmt.Fprint(w, string(b))
}

func MustError(code int, w http.ResponseWriter, format string, a ...interface{}) {
	_, err := Errorf(code, w, format, a...)
	if err != nil {
		log.WithError(err).Warn("failed to write error response")
	}
}

// ListenAndServe is the entry point for the REST API
func ListenAndServe(addr string, controllers []string) {
	iscsi, err := iscsi.New(controllers)
	if err != nil {
		log.Fatalf("Failed to initialize ISCSI: %v", err)
	}
	nfs, err := nfs.New(controllers)
	if err != nil {
		log.Fatalf("Failed to initialize NFS: %v", err)
	}
	nvmeof, err := nvmeof.New(controllers)
	if err != nil {
		log.Fatalf("Failed to initialize NVMeoF: %v", err)
	}
	s := &server{
		router: mux.NewRouter(),
		iscsi:  iscsi,
		nfs:    nfs,
		nvmeof: nvmeof,
	}

	s.routes()

	log.Fatal(http.ListenAndServe(addr, s.router))
}
