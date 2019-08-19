// Package rest provides the REST API to create highly-available iSCSI targets.
package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/LINBIT/linstor-remote-storage/crmcontrol"
	"github.com/LINBIT/linstor-remote-storage/iscsi"
	"github.com/gorilla/mux"
)

// Error is the type that is returned in case of an error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Errorf takes a StatusCode, a ResponswWriter and a format string.
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
	return fmt.Fprintf(w, string(b))
}

func unmarshalBody(w http.ResponseWriter, r *http.Request, i interface{}) error {
	var s string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s = "Could not read body"
		_, _ = Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	s = "Could not unmarshal body"
	if err := json.Unmarshal(body, i); err != nil {
		_, _ = Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	return nil
}

// ListenAndServe is the entry point for the REST API
func ListenAndServe(addr string) {
	router := mux.NewRouter() //.StrictSlash(true)
	router.HandleFunc("/api/v1/iscsi", ISCSICreate).Methods("POST")
	router.HandleFunc("/api/v1/iscsi", ISCSIDelete).Methods("DELETE")
	log.Fatal(http.ListenAndServe(addr, router))
}

func maybeSetLinstorController(iscsi *iscsi.ISCSI) {
	if iscsi.Linstor.ControllerIP == nil {
		foundIP, err := crmcontrol.FindLinstorController()
		if err == nil {
			iscsi.Linstor.ControllerIP = foundIP
		} else {
			iscsi.Linstor.ControllerIP = net.IPv4(127, 0, 0, 1)
		}
	}
}
