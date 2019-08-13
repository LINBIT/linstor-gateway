package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type restError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Errorf(code int, w http.ResponseWriter, format string, a ...interface{}) (n int, err error) {
	e := restError{
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
		Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	s = "Could not unmarshal body"
	if err := json.Unmarshal(body, i); err != nil {
		Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	return nil
}

func ListenAndServe(addr string) {
	router := mux.NewRouter() //.StrictSlash(true)
	router.HandleFunc("/api/v1/iscsi", ISCSICreate).Methods("POST")
	router.HandleFunc("/api/v1/iscsi", ISCSIDelete).Methods("DELETE")
	log.Fatal(http.ListenAndServe(addr, router))
}
