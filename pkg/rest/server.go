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
	"strconv"
	"sync"

	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/nfs"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	"github.com/gorilla/mux"
)

type server struct {
	router *mux.Router
	sync.Mutex
}

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
	return fmt.Fprint(w, string(b))
}

func unmarshalBody(w http.ResponseWriter, r *http.Request, i interface{}) error {
	var s string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s = "Could not read body: " + err.Error()
		_, _ = Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	if err := json.Unmarshal(body, i); err != nil {
		s = "Could not unmarshal body: " + err.Error()
		_, _ = Errorf(http.StatusBadRequest, w, s)
		return errors.New(s)
	}

	return nil
}

// ListenAndServe is the entry point for the REST API
func ListenAndServe(addr string) {
	s := &server{
		router: mux.NewRouter(),
	}

	s.routes()

	log.Fatal(http.ListenAndServe(addr, s.router))
}

func maybeSetLinstorController(container interface{}) {
	var linstorRsc *linstorcontrol.Linstor
	switch container.(type) {
		case *iscsi.ISCSI:
			linstorRsc = &(container.(*iscsi.ISCSI).Linstor)
		case *nfs.NFSResource:
			linstorRsc = &(container.(*nfs.NFSResource).Linstor)
		default:
			panic("Logic error: unexpected data type")
	}
	if linstorRsc.ControllerIP == nil {
		foundIP, err := crmcontrol.FindLinstorController()
		if err == nil {
			linstorRsc.ControllerIP = foundIP
		} else {
			linstorRsc.ControllerIP = net.IPv4(127, 0, 0, 1)
		}
	}
}

// parseIQNAndLun does the shared parsing for methods that are .../iscsi/{iqn}/{lun}"
func parseIQNAndLun(w http.ResponseWriter, r *http.Request) (iscsi.ISCSI, bool) {
	iscsiCfg := iscsi.ISCSI{}

	iqn, ok := mux.Vars(r)["iqn"]
	if !ok {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not find 'target' in your request")
		return iscsiCfg, false
	}

	l, ok := mux.Vars(r)["lun"]
	if !ok {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not find 'lun' in your request")
		return iscsiCfg, false
	}
	lid, err := strconv.Atoi(l)
	if err != nil {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not convert LUN to number: %v", err)
		return iscsiCfg, false
	}

	lun := targetutil.LUN{ID: uint8(lid)}
	targetConfig := targetutil.TargetConfig{
		IQN:  iqn,
		LUNs: []*targetutil.LUN{&lun},
	}
	target, err := targetutil.NewTarget(targetConfig)
	if err != nil {
		_, _ = Errorf(http.StatusInternalServerError, w, "Could not create target from target config: %v", err)
		return iscsiCfg, false
	}

	iscsiCfg.Target = target
	iscsiCfg.Linstor = linstorcontrol.Linstor{}

	maybeSetLinstorController(&iscsiCfg)

	if err := targetutil.CheckIQN(iscsiCfg.Target.IQN); err != nil {
		_, _ = Errorf(http.StatusBadRequest, w, "Could not validate IQN: %v", err)
		return iscsiCfg, false
	}
	targetName, _ := targetutil.ExtractTargetName(iscsiCfg.Target.IQN)
	iscsiCfg.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, uint8(lid))

	return iscsiCfg, true
}

// parseNFSRsc does the shared parsing for methods that are .../nfs/{resource}"
func parseNFSResource(response http.ResponseWriter, request *http.Request) (nfs.NFSResource, bool) {
	nfsRsc := nfs.NFSResource{}

	resourceName, ok := mux.Vars(request)["resource"]
	if !ok {
		_, _ = Errorf(http.StatusBadRequest, response, "The 'resource' field is absent from the URL")
		return nfsRsc, false
	}

	nfsRsc.NFS.ResourceName = resourceName
	nfsRsc.Linstor.ResourceName = resourceName

	maybeSetLinstorController(&nfsRsc)

	return nfsRsc, true
}
