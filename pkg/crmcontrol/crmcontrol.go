// Package crmcontrol provides a low-level API to manage Pacemaker.
//
// The functions in this module are called by the high-level API in package
// iscsi to perform operations in the CRM subsystem, such as creating the
// primitives and constraints that configure iSCSI targets, logical units and
// the associated service IP addresses.
// The 'cibadmin' utility is used to modify the cluster's CIB (cluster
// information base).
// The CIB is modified by
//   - sending XML entries, created from templates, to create new primitives & constraints,
//     much like a macro processor
//   - reading and parsing the current CIB XML, and modifying the contents
//     (e.g. removing tags and their nested tags) to delete existing entries from
//     the cluster configuration.
// The 'etree' package is used for XML parsing and modification.
package crmcontrol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-iscsi/pkg/crmtemplate"
	"github.com/LINBIT/linstor-iscsi/pkg/targetutil"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"

	xmltree "github.com/beevik/etree"
)

// Template variable keys
const (
	varLus     = "CRM_LUS"
	varTgtName = "CRM_TARGET_NAME"
)

// Initial delay after setting resource target-role=Stopped before starting to poll the CIB
// to check whether resources have actually stopped
const waitStopPollCibDelay = 2500

// Pacemaker CIB XML XPaths
const (
	cibRscXpath        = "/cib/configuration/resources"
	cibNodeStatusXpath = "/cib/status/node_state"
)

// CrmConfiguration stores information about (Pacemaker) CRM resources
type CrmConfiguration struct {
	Targets []*crmTarget
	LUs     []*crmLu
	IPs     []*crmIP
	TIDs    *IntSet
}

type crmTarget struct {
	ID       string
	IQN      string
	Username string
	Password string
	Portals  string
	Tid      int
}

type crmLu struct {
	ID     string
	LUN    uint8
	Target *crmTarget
	Path   string
}

type crmIP struct {
	ID      string
	IP      net.IP
	Netmask uint8
}

type ResourceRunState struct {
	TargetState cib.LrmRunState           `json:"target"`
	LUStates    map[uint8]cib.LrmRunState `json:"luns"`
	IPState     cib.LrmRunState           `json:"ip"`
}

func checkTargetExists(c *cib.CIB, iqn string) (bool, string, error) {
	targetName, err := targetutil.ExtractTargetName(iqn)
	if err != nil {
		return false, "", err
	}

	id := crmTargetID(targetName)
	elem := c.FindResource(id)
	if elem == nil {
		log.Debug("Not found")
		return false, "", nil
	}

	attr, err := cib.GetNvPairValue(elem, "iqn")
	if err != nil {
		return false, "", errors.New("could not find iqn: " + err.Error())
	}

	if iqn == attr.Value {
		log.Debug("Found")
		return true, "", nil
	}

	return false, attr.Value, nil
}

func didYouMean(iqn, suggest string) {
	log.Errorf("Unknown target %s.", aurora.Cyan(iqn))
	if suggest != "" {
		log.Errorf("Did you mean   %s?", aurora.Cyan(suggest))
	}
}

func generateCreateLuXML(target targetutil.Target, storageNodes []string,
	device string, tid int16) (string, error) {
	targetName, err := targetutil.ExtractTargetName(target.IQN)
	if err != nil {
		return "", err
	}

	tmplVars := map[string]interface{}{
		"Target":       target,
		"TargetName":   targetName,
		"StorageNodes": storageNodes,
		"Device":       device,
		"TID":          tid,
	}

	// Replace resource creation template variables
	iscsitmpl := template.Must(template.New("crmisci").Parse(crmtemplate.CRM_ISCSI))

	var cibData bytes.Buffer
	err = iscsitmpl.Execute(&cibData, tmplVars)
	return cibData.String(), err
}

// CreateCrmLu creates a CRM resource for a logical unit.
//
// The resources created depend on the contents of the template for resource creation.
// Typically, it's an iSCSI target, logical unit and service IP address, along
// with constraints that bundle them and place them on the selected nodes
func CreateCrmLu(target targetutil.Target, storageNodes []string,
	device string, tid int16) error {
	var c cib.CIB
	// Load the template for modifying the CIB
	forStdin, err := generateCreateLuXML(target, storageNodes, device, tid)
	if err != nil {
		return err
	}

	return c.CreateResource(forStdin)
}

func StartTarget(iqn string, luns []uint8) error {
	return startStopTarget(iqn, luns, true)
}

func StopTarget(iqn string, luns []uint8) error {
	return startStopTarget(iqn, luns, false)
}

func startStopTarget(iqn string, luns []uint8, start bool) error {
	var c cib.CIB
	// Read the current CIB XML
	err := c.ReadConfiguration()
	if err != nil {
		return err
	}

	exists, suggest, err := checkTargetExists(&c, iqn)
	if err != nil {
		return err
	}
	if !exists {
		didYouMean(iqn, suggest)
		if start {
			return errors.New("Unable to start resource")
		} else {
			return errors.New("Unable to stop resource")
		}
	}

	target, err := targetutil.ExtractTargetName(iqn)
	if err != nil {
		return err
	}

	ids := generateCrmObjectNames(target, luns)
	for _, id := range ids {
		if start {
			err = c.StartResource(id)
		} else {
			err = c.StopResource(id)
		}
		if err != nil {
			return err
		}
	}

	return c.Update()
}

// getIDsToDelete figures out what CRM objects need to be deleted given a LUN.
func getIDsToDelete(c *cib.CIB, target string, lun uint8) ([]string, error) {
	// Count LUNs in the cluster which belong to this target
	numLuns := 0
	lunElems := c.Doc.FindElements("cib/configuration/resources/primitive[@type='iSCSILogicalUnit']")
	for _, lunElem := range lunElems {
		idAttr := lunElem.SelectAttr("id")
		if idAttr == nil {
			log.WithFields(log.Fields{
				"target": target,
			}).Warning("CRM iSCSILogicalUnit without id")
			continue
		}

		regexBelongs := `^` + crmTargetID(target) + `_lu\d+$`
		matched, err := regexp.MatchString(regexBelongs, idAttr.Value)
		if err != nil {
			return nil, err
		} else if !matched {
			log.WithFields(log.Fields{
				"target": target,
				"lu":     idAttr.Value,
			}).Debug("LU does not seem to belong to target, skipping.")
			continue
		}

		numLuns++
	}

	if numLuns == 0 {
		return []string{}, nil
	} else if numLuns == 1 {
		// this was the only LU -> delete everything related to this target
		return generateCrmObjectNames(target, []uint8{lun}), nil
	} else {
		// there are still more LUs -> only delete this one
		return []string{crmLuID(target, lun)}, nil
	}
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// DeleteLogicalUnit deletes the CRM resources for a target
func DeleteLogicalUnit(iqn string, lun uint8) error {
	var c cib.CIB
	// Read the current CIB XML
	err := c.ReadConfiguration()
	if err != nil {
		return err
	}

	exists, suggest, err := checkTargetExists(&c, iqn)
	if err != nil {
		return err
	}
	if !exists {
		didYouMean(iqn, suggest)
		return errors.New("Unable to delete resource")
	}

	iscsiTargetName, err := targetutil.ExtractTargetName(iqn)
	if err != nil {
		return err
	}

	if c.FindResource(crmLuID(iscsiTargetName, lun)) == nil {
		return fmt.Errorf("Target %s does not have LUN %d", aurora.Cyan(iscsiTargetName), aurora.Cyan(lun))
	}

	ids, err := getIDsToDelete(&c, iscsiTargetName, lun)
	if err != nil {
		return err
	}

	allPresent := true
	for _, id := range ids {
		if c.FindResource(id) == nil {
			log.WithFields(log.Fields{
				"id": id,
			}).Debug("ID not found, not all resources are present")
			allPresent = false
			ids = remove(ids, id)
		}
	}

	if !allPresent {
		log.Warning("Partial resource state detected. Deleting remaining CRM resources.")
	}

	log.Debug("Deleting these IDs:")
	for _, id := range ids {
		log.Debugf("    %s", id)
	}

	// notify pacemaker to delete the IDs
	for _, id := range ids {
		err = c.StopResource(id)
		if err != nil {
			log.WithFields(log.Fields{
				"resource": id,
			}).Warning("Could not set target-role. Resource will probably fail to stop: ", err)
		}
	}
	err = c.Update()
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(waitStopPollCibDelay * time.Millisecond))
	isStopped, err := c.WaitForResourcesStop(ids)
	if err != nil {
		return err
	}

	if !isStopped {
		return errors.New("Resource stop was not confirmed for all resources, cannot continue delete action")
	}

	// Read the current CIB XML
	err = c.ReadConfiguration()
	if err != nil {
		return err
	}

	// Process the CIB XML document tree, removing constraints that refer to any of the objects
	// that will be deleted
	c.DissolveConstraints(ids)

	// Process the CIB XML document tree, removing the specified CRM resources
	for _, id := range ids {
		rscElem := c.FindResource(id)
		if rscElem != nil {
			rscElemParent := rscElem.Parent()
			if rscElemParent != nil {
				rscElemParent.RemoveChildAt(rscElem.Index())
			} else {
				return errors.New("Cannot modify CIB, CRM resource '" + id + "' has no parent object")
			}
		} else {
			fmt.Printf("Warning: CRM resource '%s' not found in the CIB\n", id)
		}
	}

	return c.Update()
}

func crmTargetID(target string) string {
	return "p_iscsi_" + target
}

func crmLuID(target string, lun uint8) string {
	return "p_iscsi_" + target + "_lu" + strconv.Itoa(int(lun))
}

func crmIPID(target string) string {
	return "p_iscsi_" + target + "_ip"
}

// ProbeResource probes the LRM run state of the CRM resources associated with the specified iSCSI resource
func ProbeResource(iqn string, luns []uint8) (ResourceRunState, error) {
	state := ResourceRunState{
		TargetState: cib.Unknown,
		LUStates:    make(map[uint8]cib.LrmRunState),
		IPState:     cib.Unknown,
	}

	var c cib.CIB

	// Read the current CIB XML
	err := c.ReadConfiguration()
	if err != nil {
		return state, err
	}

	exists, suggest, err := checkTargetExists(&c, iqn)
	if err != nil {
		return state, err
	}
	if !exists {
		didYouMean(iqn, suggest)
		return state, errors.New("Unable to probe resource")
	}

	target, err := targetutil.ExtractTargetName(iqn)
	if err != nil {
		return state, err
	}

	state.TargetState = c.FindLrmState(crmTargetID(target))
	for _, lun := range luns {
		state.LUStates[lun] = c.FindLrmState(crmLuID(target, lun))
	}
	state.IPState = c.FindLrmState(crmIPID(target))

	return state, nil
}

func (c *CrmConfiguration) findTargetByIqn(iqn string) (*crmTarget, error) {
	for _, t := range c.Targets {
		if t.IQN == iqn {
			return t, nil
		}
	}

	return nil, errors.New("no target with IQN found")
}

func findTargets(rscSection *xmltree.Element) []*crmTarget {
	targets := make([]*crmTarget, 0)
	for _, target := range rscSection.FindElements("./primitive[@type='iSCSITarget']") {
		// find ID
		id := target.SelectAttr("id")
		if id == nil {
			log.Debug("Skipping invalid iSCSITarget without id")
			continue
		}

		contextLog := log.WithFields(log.Fields{"id": id.Value})

		// find IQN
		iqn, err := cib.GetNvPairValue(target, "iqn")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without iqn: ", err)
			continue
		}

		contextLog = log.WithFields(log.Fields{"id": id.Value, "iqn": iqn.Value})

		// find username
		username, err := cib.GetNvPairValue(target, "incoming_username")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without username: ", err)
			continue
		}

		// find password
		password, err := cib.GetNvPairValue(target, "incoming_password")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without password: ", err)
			continue
		}

		// find portals
		portals, err := cib.GetNvPairValue(target, "portals")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without portals: ", err)
			continue
		}

		// find tid
		tidAttr, err := cib.GetNvPairValue(target, "tid")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without TID: ", err)
			continue
		}

		tid, err := strconv.Atoi(tidAttr.Value)
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget with invalid LUN: ", err)
			continue
		}

		crmTarget := &crmTarget{
			ID:       id.Value,
			IQN:      iqn.Value,
			Username: username.Value,
			Password: password.Value,
			Portals:  portals.Value,
			Tid:      tid,
		}

		targets = append(targets, crmTarget)
	}
	return targets
}

func findLus(rscSection *xmltree.Element, config *CrmConfiguration) []*crmLu {
	lus := make([]*crmLu, 0)
	for _, lu := range rscSection.FindElements("./primitive[@type='iSCSILogicalUnit']") {
		// find ID
		id := lu.SelectAttr("id")
		if id == nil {
			log.Debug("Skipping invalid iSCSILogicalUnit without id")
			continue
		}

		contextLog := log.WithFields(log.Fields{"id": id.Value})

		// find LUN
		lunAttr, err := cib.GetNvPairValue(lu, "lun")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSILogicalUnit without LUN: ", err)
			continue
		}

		contextLog = log.WithFields(log.Fields{"id": id.Value, "lun": lunAttr.Value})

		lun, err := strconv.ParseInt(lunAttr.Value, 10, 8)
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSILogicalUnit with invalid LUN: ", err)
			continue
		}

		contextLog = log.WithFields(log.Fields{"id": id.Value, "lun": lun})

		// find target IQN
		targetIqn, err := cib.GetNvPairValue(lu, "target_iqn")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSILogicalUnit without target iqn: ", err)
			continue
		}

		contextLog = log.WithFields(log.Fields{
			"id":     id.Value,
			"lun":    lun,
			"target": targetIqn.Value,
		})

		// find associated target
		target, err := config.findTargetByIqn(targetIqn.Value)
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSILogicalUnit with unknown target: ", err)
			continue
		}

		// find path
		path, err := cib.GetNvPairValue(lu, "path")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSILogicalUnit without path: ", err)
			continue
		}

		crmLu := &crmLu{
			ID:     id.Value,
			LUN:    uint8(lun),
			Target: target,
			Path:   path.Value,
		}

		lus = append(lus, crmLu)
	}

	return lus
}

func findIPs(rscSection *xmltree.Element) []*crmIP {
	ips := make([]*crmIP, 0)
	for _, ipElem := range rscSection.FindElements("./primitive[@type='IPaddr2']") {
		// find ID
		id := ipElem.SelectAttr("id")
		if id == nil {
			log.Debug("Skipping invalid IPaddr2 without id")
			continue
		}

		contextLog := log.WithFields(log.Fields{"id": id.Value})

		// find ip
		ipAddr, err := cib.GetNvPairValue(ipElem, "ip")
		if err != nil {
			contextLog.Debug("Skipping invalid IPaddr2 without ip: ", err)
			continue
		}

		// find netmask
		netmaskAttr, err := cib.GetNvPairValue(ipElem, "cidr_netmask")
		if err != nil {
			contextLog.Debug("Skipping invalid IPaddr2 without netmask: ", err)
			continue
		}

		netmask, err := strconv.ParseInt(netmaskAttr.Value, 10, 8)
		if err != nil {
			contextLog.Debug("Skipping invalid IPaddr2 with invalid netmask: ", err)
			continue
		}

		ip := &crmIP{
			ID:      id.Value,
			IP:      net.ParseIP(ipAddr.Value),
			Netmask: uint8(netmask),
		}

		ips = append(ips, ip)
	}

	return ips
}

// ParseConfiguration parses the CIB XML document and returns information about
// existing resources.
//
// Information about existing CRM resources is parsed from the CIB XML document and
// stored in a newly allocated crmConfiguration data structure
// TODO THINK: maybe we can replace this whole mess by actual standard Go XML marshalling...
func ParseConfiguration(docRoot *xmltree.Document) (*CrmConfiguration, error) {
	config := CrmConfiguration{TIDs: NewIntSet()}
	if docRoot == nil {
		return nil, errors.New("Internal error: ParseConfiguration() called with docRoot == nil")
	}

	cib := docRoot.Root()
	if cib == nil {
		return nil, errors.New("Failed to find the cluster information base (CIB) root element")
	}

	rscSection := cib.FindElement(cibRscXpath)
	if rscSection == nil {
		return nil, errors.New("Failed to find the cluster resources section in the cluster information base (CIB)")
	}

	config.Targets = findTargets(rscSection)
	config.LUs = findLus(rscSection, &config)
	config.IPs = findIPs(rscSection)

	return &config, nil
}

// generateCrmObjectNames generates a list of all CRM object names for a target
func generateCrmObjectNames(iscsiTargetName string, luns []uint8) []string {
	objects := make([]string, 0)

	templateVars := make(map[string]interface{})
	templateVars[varTgtName] = iscsiTargetName
	templateVars[varLus] = luns

	tmpl := template.Must(template.New("crmobjnames").Parse(crmtemplate.CRM_OBJ_NAMES))

	var buf bytes.Buffer
	tmpl.Execute(&buf, templateVars)

	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		name := strings.TrimRight(scanner.Text(), "\r\n")
		log.Debug("crm object name: ", name)
		objects = append(objects, name)
	}
	return objects
}
