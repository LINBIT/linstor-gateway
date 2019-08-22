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

	"github.com/LINBIT/linstor-iscsi/pkg/crmtemplate"
	log "github.com/sirupsen/logrus"

	xmltree "github.com/beevik/etree"
)

// Template variable keys
const (
	varNodeName    = "CRM_NODE_NAME"
	varNr          = "NR"
	varLuName      = "CRM_LU_NAME"
	varLus         = "CRM_LUS"
	varSvcIP       = "CRM_SVC_IP"
	varTgtName     = "CRM_TARGET_NAME"
	varTgtIqn      = "TARGET_IQN"
	varIscsiLun    = "LUN"
	varStorDev     = "DEVICE"
	varUsername    = "USERNAME"
	varPassword    = "PASSWORD"
	varPortals     = "PORTALS"
	varTid         = "TID"
	varTgtLocNodes = "TARGET_LOCATION_NODES"
	varLuLocNodes  = "LU_LOCATION_NODES"
)

// Pacemaker CIB XML XPaths
const (
	cibRscXpath        = "/cib/configuration/resources"
	cibStatusXpath     = "/cib/status"
	cibNodeStatusXpath = "/cib/status/node_state"
)

// Pacemaker CRM resource names, prefixes, suffixes, search patterns, etc.
const (
	crmIscsiRscPrefix  = "p_iscsi_"
	crmIscsiLuName     = "lu"
	crmIscsiPrmTid     = "tid"
	crmTypeIscsiTarget = "iSCSITarget"
	crmTypeIscsiLu     = "iSCSILogicalUnit"
	crmTypeLinstorCtrl = "linstor-controller"
)

// Pacemaker CIB XML tag names
const (
	cibTagLocation   = "rsc_location"
	cibTagColocation = "rsc_colocation"
	cibTagOrder      = "rsc_order"
	cibTagRscRef     = "resource_ref"
	cibTagMetaAttr   = "meta_attributes"
	cibTagInstAttr   = "instance_attributes"
	cibTagNvPair     = "nvpair"
	cibTagLrm        = "lrm"
	cibTagLrmRsclist = "lrm_resources"
	cibTagLrmRsc     = "lrm_resource"
	cibTagLrmRscOp   = "lrm_rsc_op"
)

// Pacemaker CIB attribute names
const (
	cibAttrKeyID           = "id"
	cibAttrKeyName         = "name"
	cibAttrKeyValue        = "value"
	cibAttrKeyOperation    = "operation"
	cibAttrKeyRcCode       = "rc-code"
	cibAttrValueTargetRole = "target-role"
	cibAttrValueStarted    = "Started"
	cibAttrValueStopped    = "Stopped"
	cibAttrValueStop       = "stop"
	cibAttrValueStart      = "start"
	cibAttrValueMonitor    = "monitor"
)

// Pacemaker OCF resource agent exit codes
const (
	ocfSuccess          = 0
	ocfErrGeneric       = 1
	ocfErrArgs          = 2
	ocfErrUnimplemented = 3
	ocfErrPerm          = 4
	ocfErrInstalled     = 5
	ocfNotRunning       = 7
	ocfRunningMaster    = 8
	ocfFailedMaster     = 9
)

// Maximum recursion level, currently used to limit recursion during recursive
// searches of the XML document tree
const maxRecursionLevel = 40

// Maximum number of CIB poll retries when waiting for CRM resources to stop
const maxWaitStopRetries = 10

// Initial delay after setting resource target-role=Stopped before starting to poll the CIB
// to check whether resources have actually stopped
const waitStopPollCibDelay = 2500

// Delay between CIB polls in milliseconds
const cibPollRetryDelay = 2000

var errMaxRecursion = errors.New("Exceeding maximum recursion level, operation aborted")

// CrmConfiguration stores information about (Pacemaker) CRM resources
type CrmConfiguration struct {
	TargetList   []*crmTarget
	LuList       []*crmLu
	IPList       []*crmIP
	OtherRscList []string
	TidSet       *IntSet
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
	TargetState LrmRunState
	LUStates    map[uint8]LrmRunState
	IPState     LrmRunState
}

// LrmRunState represents the state of a CRM resource.
type LrmRunState int

const (
	// Unknown means that the resource's state could not be retrieved
	Unknown LrmRunState = iota
	// Running means that the resource is verified as running
	Running
	// Stopped means that the resource is verfied as stopped
	Stopped
)

// CreateCrmLu creates a CRM resource for a logical unit.
//
// The resources created depend on the contents of the template for resource creation.
// Typically, it's an iSCSI target, logical unit and service IP address, along
// with constraints that bundle them and place them on the selected nodes
func CreateCrmLu(
	storageNodeList []string,
	iscsiTargetName string,
	ip net.IP,
	iscsiTargetIqn string,
	lun uint8,
	device string,
	username string,
	password string,
	portal string,
	tid int16,
) error {
	// Load the template for modifying the CIB

	// debug.PrintfLnCaption("Template input:")
	// debug.PrintTextArray(tmplLines)

	// Construct the CIB update data from the template
	tmplVars := make(map[string]string)
	tmplVars[varLuName] = "lu" + strconv.Itoa(int(lun))
	tmplVars[varSvcIP] = ip.String()
	tmplVars[varTgtName] = iscsiTargetName
	tmplVars[varTgtIqn] = iscsiTargetIqn
	tmplVars[varIscsiLun] = strconv.Itoa(int(lun))
	tmplVars[varStorDev] = device
	tmplVars[varUsername] = username
	tmplVars[varPassword] = password
	tmplVars[varPortals] = portal
	tmplVars[varTid] = strconv.Itoa(int(tid))

	// Create sub XML content, one entry per node, from the iSCSI target location constraint template
	targetLocData, err := constructNodesTemplate(crmtemplate.TARGET_LOCATION_NODES, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	// Create sub XML content, one entry per node, from the iSCSI logical unit location constraint template
	luLocData, err := constructNodesTemplate(crmtemplate.LU_LOCATION_NODES, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	// Load the sub XML content into variables
	tmplVars[varTgtLocNodes] = targetLocData
	tmplVars[varLuLocNodes] = luLocData

	// Replace resource creation template variables
	iscsitmpl, err := template.New("crmisci").Parse(crmtemplate.CRM_ISCSI)
	if err != nil {
		return err
	}
	var cibData bytes.Buffer
	iscsitmpl.Execute(&cibData, tmplVars)

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	forStdin := cibData.String()
	stdout, stderr, err := execute(&forStdin, crmCreateCommand.executable, crmCreateCommand.arguments...)
	if err != nil {
		return err
	}

	if len(stdout) >= 1 {
		log.Debug("Begin of CRM command stdout output:", stdout)
	} else {
		log.Debug("No stdout output")
	}

	if len(stderr) >= 1 {
		log.Debug("CRM command stderr output:", stderr)
	} else {
		log.Debug("No stdout output")
	}

	return err
}

// ModifyCrmTargetRole sets the target-role of a resource in CRM.
//
// The id has to be a valid CRM resource identifier.
// A target-role of "Stopped" (startFlag == false) indicates to CRM that
// the it should stop the resource. A target role of "Started" (startFlag == true)
// indicates that the resource is already started and that CRM should not try
// to start it.
func ModifyCrmTargetRole(id string, startFlag bool, doc *xmltree.Document) (*xmltree.Document, error) {
	// Process the CIB XML document tree and insert meta attributes for target-role=Stopped
	rscElem := doc.FindElement("/cib/configuration/resources/primitive[@id='" + id + "']")
	if rscElem == nil {
		return nil, errors.New("CRM resource not found in the CIB, cannot modify role.")
	}

	var tgtRoleEntry *xmltree.Element
	metaAttr := rscElem.FindElement(cibTagMetaAttr)
	if metaAttr != nil {
		// Meta attributes exist, find the entry that sets the target-role
		tgtRoleEntry = metaAttr.FindElement(cibTagNvPair + "[@" + cibAttrKeyName + "='" + cibAttrValueTargetRole + "']")
	} else {
		// No meta attributes present, create XML element
		metaAttr = rscElem.CreateElement(cibTagMetaAttr)
		metaAttr.CreateAttr(cibAttrKeyID, id+"_"+cibTagMetaAttr)
	}
	if tgtRoleEntry == nil {
		// No entry that sets the target-role, create entry
		tgtRoleEntry = metaAttr.CreateElement(cibTagNvPair)
		tgtRoleEntry.CreateAttr(cibAttrKeyID, id+"_"+cibAttrValueTargetRole)
		tgtRoleEntry.CreateAttr(cibAttrKeyName, cibAttrValueTargetRole)
	}
	// Set the target-role
	var tgtRoleValue string
	if startFlag {
		tgtRoleValue = cibAttrValueStarted
	} else {
		tgtRoleValue = cibAttrValueStopped
	}
	tgtRoleEntry.CreateAttr(cibAttrKeyValue, tgtRoleValue)

	return doc, nil
}

func StartCrmResource(target string, luns []uint8) error {
	// Read the current CIB XML
	doc, err := ReadConfiguration()
	if err != nil {
		return err
	}

	log.Debugf("starting target %s LUNs %v", target, luns)

	ids := generateCrmObjectNames(target, luns)
	for _, id := range ids {
		doc, err = ModifyCrmTargetRole(id, true, doc)
		if err != nil {
			return err
		}
	}

	return executeCibUpdate(doc, crmUpdateCommand)
}

func StopCrmResource(target string, luns []uint8) error {
	// Read the current CIB XML
	doc, err := ReadConfiguration()
	if err != nil {
		return err
	}

	ids := generateCrmObjectNames(target, luns)
	for _, id := range ids {
		doc, err = ModifyCrmTargetRole(id, false, doc)
		if err != nil {
			return err
		}
	}

	return executeCibUpdate(doc, crmUpdateCommand)
}

// getIDsToDelete figures out what CRM objects need to be deleted given a LUN.
func getIDsToDelete(target string, lun uint8) ([]string, error) {
	// Read the current CIB XML
	doc, err := ReadConfiguration()
	if err != nil {
		return nil, err
	}

	// Count LUNs in the cluster which belong to this target
	numLuns := 0
	lunElems := doc.FindElements("cib/configuration/resources/primitive[@type='iSCSILogicalUnit']")
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
		return nil, errors.New("Logic error: there should be at least one Logical Unit for this target")
	} else if numLuns == 1 {
		// this was the only LU -> delete everything related to this target
		return generateCrmObjectNames(target, []uint8{lun}), nil
	} else {
		// there are still more LUs -> only delete this one
		return []string{crmLuID(target, lun)}, nil
	}
}

// DeleteCrmLu deletes the CRM resources for a target
func DeleteCrmLu(iscsiTargetName string, lun uint8) error {
	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return err
	}

	ids, err := getIDsToDelete(iscsiTargetName, lun)
	if err != nil {
		return err
	}

	log.Debug("Deleting these IDs:")
	for _, id := range ids {
		log.Debugf("    %s", id)
	}

	// notify pacemaker to delete the IDs
	for _, id := range ids {
		docRoot, err = ModifyCrmTargetRole(id, false, docRoot)
		if err != nil {
			log.WithFields(log.Fields{
				"resource": id,
			}).Warning("Could not set target-role. Resource will probably fail to stop.")
		}
	}
	err = executeCibUpdate(docRoot, crmUpdateCommand)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(waitStopPollCibDelay * time.Millisecond))
	isStopped, err := waitForResourcesStop(ids)
	if err != nil {
		return err
	}

	if !isStopped {
		return errors.New("Resource stop was not confirmed for all resources, cannot continue delete action")
	}

	// Read the current CIB XML
	docRoot, err = ReadConfiguration()
	if err != nil {
		return err
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	// Process the CIB XML document tree, removing constraints that refer to any of the objects
	// that will be deleted
	err = dissolveConstraints(cib, ids)
	if err != nil {
		return err
	}

	// Process the CIB XML document tree, removing the specified CRM resources
	for _, elemID := range ids {
		rscElem := cib.FindElement("/cib/configuration/resources/primitive[@id='" + elemID + "']")
		if rscElem != nil {
			rscElemParent := rscElem.Parent()
			if rscElemParent != nil {
				rscElemParent.RemoveChildAt(rscElem.Index())
			} else {
				return errors.New("Cannot modify CIB, CRM resource '" + elemID + "' has no parent object")
			}
		} else {
			fmt.Printf("Warning: CRM resource '%s' not found in the CIB\n", elemID)
		}
	}

	return executeCibUpdate(docRoot, crmUpdateCommand)
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
func ProbeResource(target string, luns []uint8) (ResourceRunState, error) {
	state := ResourceRunState{
		TargetState: Unknown,
		LUStates:    make(map[uint8]LrmRunState),
		IPState:     Unknown,
	}

	// Read the current CIB XML
	doc, err := ReadConfiguration()
	if err != nil {
		return state, err
	}

	state.TargetState = findLrmState(crmTargetID(target), doc)
	for _, lun := range luns {
		state.LUStates[lun] = findLrmState(crmLuID(target, lun), doc)
	}
	state.IPState = findLrmState(crmIPID(target), doc)

	return state, nil
}

func resourceInCIB(docRoot *xmltree.Document, id string) bool {
	return docRoot.FindElement("//primitive[@id='"+id+"']") != nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// waitForResourcesStop waits for CRM resources to stop
//
// It returns a flag indicating whether resources are stopped (true) or
// not (false), and an error.
func waitForResourcesStop(idsToStop []string) (bool, error) {
	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return false, err
	}

	for _, id := range idsToStop {
		if !resourceInCIB(docRoot, id) {
			log.WithFields(log.Fields{
				"resource": id,
			}).Warning("Resource not found in the CIB, will be ignored.")
			idsToStop = remove(idsToStop, id)
		}
	}

	log.Debug("Waiting for the following CRM resources to stop:")
	for _, id := range idsToStop {
		log.Debugf("    %s", id)
	}

	isStopped := false
	retries := 0
	for !isStopped {
		// check if all resources are stopped
		allStopped := true
		for _, item := range idsToStop {
			state := findLrmState(item, docRoot)
			if state != Stopped {
				allStopped = false
				break
			}
		}

		if allStopped {
			// success; we stopped all resources
			isStopped = true
			break
		}

		if retries > maxWaitStopRetries {
			// timeout
			isStopped = false
			break
		}

		time.Sleep(time.Duration(cibPollRetryDelay * time.Millisecond))

		// Re-read the current CIB XML
		docRoot, err = ReadConfiguration()
		if err != nil {
			return false, err
		}

		retries++
	}

	if isStopped {
		log.Debug("The resources are stopped")
	} else {
		log.Warning("Could not confirm that the resources are stopped")
	}

	return isStopped, nil
}

func getNvPairValue(elem *xmltree.Element, name string) (*xmltree.Attr, error) {
	xpath := fmt.Sprintf("./instance_attributes/nvpair[@name='%s']", name)

	var nvpair *xmltree.Element
	if nvpair = elem.FindElement(xpath); nvpair == nil {
		return nil, errors.New("key not found")
	}

	var attr *xmltree.Attr
	if attr = nvpair.SelectAttr("value"); attr == nil {
		return nil, errors.New("value not found")
	}

	return attr, nil
}

func (c *CrmConfiguration) findTargetByIqn(iqn string) (*crmTarget, error) {
	for _, t := range c.TargetList {
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
		iqn, err := getNvPairValue(target, "iqn")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without iqn: ", err)
			continue
		}

		contextLog = log.WithFields(log.Fields{"id": id.Value, "iqn": iqn.Value})

		// find username
		username, err := getNvPairValue(target, "incoming_username")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without username: ", err)
			continue
		}

		// find password
		password, err := getNvPairValue(target, "incoming_password")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without username: ", err)
			continue
		}

		// find portals
		portals, err := getNvPairValue(target, "portals")
		if err != nil {
			contextLog.Debug("Skipping invalid iSCSITarget without portals: ", err)
			continue
		}

		// find tid
		tidAttr, err := getNvPairValue(target, "tid")
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
		lunAttr, err := getNvPairValue(lu, "lun")
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
		targetIqn, err := getNvPairValue(lu, "target_iqn")
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
		path, err := getNvPairValue(lu, "path")
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
		ipAddr, err := getNvPairValue(ipElem, "ip")
		if err != nil {
			contextLog.Debug("Skipping invalid IPaddr2 without ip: ", err)
			continue
		}

		// find netmask
		netmaskAttr, err := getNvPairValue(ipElem, "cidr_netmask")
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
	config := CrmConfiguration{TidSet: NewIntSet()}
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

	config.TargetList = findTargets(rscSection)
	config.LuList = findLus(rscSection, &config)
	config.IPList = findIPs(rscSection)

	return &config, nil
}

// ReadConfiguration calls the crm list command and parses the XML data it returns.
func ReadConfiguration() (*xmltree.Document, error) {
	stdout, stderr, err := execute(nil, crmListCommand.executable, crmListCommand.arguments...)
	if err != nil {
		return nil, err
	}
	if len(stderr) > 0 {
		log.Debug("External command error output:", stderr)
	}

	docRoot := xmltree.NewDocument()
	err = docRoot.ReadFromString(stdout)
	if err != nil {
		return nil, err
	}

	return docRoot, nil
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

func findLrmState(id string, doc *xmltree.Document) LrmRunState {
	state := Unknown
	xpath := "cib/status/node_state/lrm/lrm_resources/lrm_resource[@id='" + id + "']"
	elems := doc.FindElements(xpath)
	for _, elem := range elems {
		state = updateRunState(id, elem, state)
	}

	return state
}

func checkResourceStopped(stopItems map[string]LrmRunState) (bool, bool) {
	stateCtr := 0
	stoppedCtr := 0
	for name, state := range stopItems {
		contextLog := log.WithFields(log.Fields{"resource": name})
		if state == Unknown {
			contextLog.Warning("No run status information for resource")
			continue
		}

		stateCtr++
		if state == Running {
			contextLog.Debug("Resource is running")
		} else {
			contextLog.Debug("Resource is stopped")
			stoppedCtr++
		}
	}

	haveState := stateCtr == len(stopItems)
	stoppedFlag := stoppedCtr == len(stopItems)

	return haveState, stoppedFlag
}

func executeCibUpdate(docRoot *xmltree.Document, crmCmd crmCommand) error {
	// Serialize the modified XML document tree into a string containing the XML document (CIB update data)
	cibData, err := docRoot.WriteToString()
	if err != nil {
		return err
	}

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	stdout, stderr, err := execute(&cibData, crmCmd.executable, crmCmd.arguments...)
	if err != nil {
		log.Warn("CRM command execution returned an error")
		log.Trace("The updated CIB data sent to the command was:")
		log.Trace(cibData)
	}

	if len(stdout) >= 1 {
		log.Debug("Begin of CRM command stdout output:", stdout)
	} else {
		log.Debug("No stdout output\n")
	}

	if len(stderr) >= 1 {
		log.Debug("CRM command stderr output:", stderr)
	} else {
		log.Debug("No stderr output")
	}

	return err
}

// Creates and returns a copy of a map[string]string
func copyMap(srcMap map[string]string) map[string]string {
	resultMap := make(map[string]string, len(srcMap))
	for key, value := range srcMap {
		resultMap[key] = value
	}
	return resultMap
}

// Constructs a sub template for each node entry
//
// For each node entry, a template is loaded and variable replacement is performed, with one of the variables
// containing the node name for the current iteration. The templates are concatenated.
// The resulting XML content is a sub template for insertion into another XML template.
func constructNodesTemplate(tmplString string, nodeList []string, tmplVars map[string]string) (string, error) {
	subTmplVars := copyMap(tmplVars)
	nr := 0
	var subDataBld strings.Builder
	for _, nodename := range nodeList {
		subTmplVars[varNodeName] = nodename
		subTmplVars[varNr] = strconv.FormatUint(uint64(nr), 10)

		tmpl, err := template.New(nodename).Parse(tmplString)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		tmpl.Execute(&buf, subTmplVars)

		subDataBld.WriteString(buf.String())
		nr++
	}
	return subDataBld.String(), nil
}

// Removes CRM constraints that refer to the specified delItems names from the CIB XML document tree
func dissolveConstraints(cibElem *xmltree.Element, delItems []string) error {
	return dissolveConstraintsImpl(cibElem, delItems, 0)
}

// See dissolveConstraints(...)
func dissolveConstraintsImpl(cibElem *xmltree.Element, delItems []string, recursionLevel int) error {
	// delIdxSet is allocated on-demand only if it is required
	var delIdxSet *IntSet

	childList := cibElem.ChildElements()
	for _, subElem := range childList {
		dependFlag := false
		var err error
		if subElem.Tag == cibTagColocation {
			if recursionLevel < maxRecursionLevel {
				dependFlag, err = isColocationDependency(subElem, delItems)
				if err != nil {
					return err
				}
				if !dependFlag {
					dependFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
					if err != nil {
						return err
					}
				}
			} else {
				return errMaxRecursion
			}
		} else if subElem.Tag == cibTagOrder {
			if recursionLevel < maxRecursionLevel {
				dependFlag, err = isOrderDependency(subElem, delItems)
				if err != nil {
					return err
				}
				if !dependFlag {
					dependFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
					if err != nil {
						return err
					}
				}
			} else {
				return errMaxRecursion
			}
		} else if subElem.Tag == cibTagLocation {
			if recursionLevel < maxRecursionLevel {
				dependFlag = isLocationDependency(subElem, delItems)
				if !dependFlag {
					dependFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
					if err != nil {
						return err
					}
				}
			} else {
				return errMaxRecursion
			}
		} else if subElem.Tag == cibTagLrmRsc {
			if recursionLevel < maxRecursionLevel {
				dependFlag, err = isLrmDependency(subElem, delItems)
				if err != nil {
					return err
				}
			} else {
				return errMaxRecursion
			}
		} else {
			if recursionLevel < maxRecursionLevel {
				err := dissolveConstraintsImpl(subElem, delItems, recursionLevel+1)
				if err != nil {
					return err
				}
			} else {
				return errMaxRecursion
			}
		}
		if dependFlag {
			if delIdxSet == nil {
				delIdxSet = NewIntSet()
			}
			delIdxSet.Add(subElem.Index())
			idAttr := subElem.SelectAttr("id")
			if idAttr != nil {
				log.WithFields(log.Fields{
					"type": subElem.Tag,
					"id":   idAttr.Value,
				}).Debug("Deleting dependency")
			}
		}
	}
	// Elements are deleted in order of descending index, so that the index of elements
	// deleted later does not change due to reordering elements that had a greater index
	// than an element thas was deleted from the slice/array.
	if delIdxSet != nil {
		for _, delIdx := range delIdxSet.ReverseSortedKeys() {
			cibElem.RemoveChildAt(delIdx)
		}
	}

	return nil
}

// Indicates whether an element has sub elements that are resource reference tags that refer to any of the specified delItems names
func hasRscRefDependency(cibElem *xmltree.Element, delItems []string, recursionLevel int) (bool, error) {
	depFlag := false

	var err error
	childList := cibElem.ChildElements()
	for _, subElem := range childList {
		if subElem.Tag == cibTagRscRef {
			idAttr := subElem.SelectAttr("id")
			if idAttr != nil {
				for _, s := range delItems {
					if s == idAttr.Value {
						depFlag = true
						break
					}
				}
			} else {
				return false, errors.New("Unparseable " + subElem.Tag + " tag, cannot find \"id\" attribute")
			}
		} else {
			if recursionLevel < maxRecursionLevel {
				depFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
				if err != nil {
					return false, err
				}
			} else {
				return false, errMaxRecursion
			}
		}
		if depFlag {
			break
		}
	}

	return depFlag, nil
}

// Indicates whether the element is a CRM location constraint that refers to any of the specified delItems names
func isLocationDependency(cibElem *xmltree.Element, delItems []string) bool {
	rscAttr := cibElem.SelectAttr("rsc")
	if rscAttr == nil {
		return false
	}

	for _, s := range delItems {
		if s == rscAttr.Value {
			return true
		}
	}

	return false
}

// Indicates whether the element is a CRM order constraint that refers to any of the specified delItems names
func isOrderDependency(cibElem *xmltree.Element, delItems []string) (bool, error) {
	firstAttr := cibElem.SelectAttr("first")
	thenAttr := cibElem.SelectAttr("then")
	if firstAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"first\" attribute")
	}
	if thenAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"then\" attribute")
	}

	for _, s := range delItems {
		if s == firstAttr.Value || s == thenAttr.Value {
			return true, nil
		}
	}

	return false, nil
}

// Indicates whether the element is a CRM colocation constraint that refers to any of the specified delItems names
func isColocationDependency(cibElem *xmltree.Element, delItems []string) (bool, error) {
	rscAttr := cibElem.SelectAttr("rsc")
	withRscAttr := cibElem.SelectAttr("with-rsc")
	if rscAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"rsc\" attribute")
	}
	if withRscAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"with-rsc\" attribute")
	}

	for _, s := range delItems {
		if s == rscAttr.Value || s == withRscAttr.Value {
			return true, nil
		}
	}

	return false, nil
}

// Indicates whether the element is an LRM entry that refers to any of the specified delItems names
func isLrmDependency(cibElem *xmltree.Element, delItems []string) (bool, error) {
	idAttr := cibElem.SelectAttr("id")
	if idAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"id\" attribute")
	}

	for _, s := range delItems {
		if s == idAttr.Value {
			return true, nil
		}
	}

	return false, nil
}

// updateRunState updates the run state information of a single resource
//
// For a resource to be considered stopped, this function must find
// - either a successful stop action
// - or a monitor action with rc-code ocfNotRunning and no stop action
//
// If a stop action is present, the monitor action can still show "running"
// (rc-code ocfSuccess == 0) although the resource is actually stopped. The
// monitor action's rc-code is only interesting if there is no stop action present.
func updateRunState(rscName string, lrmRsc *xmltree.Element, runState LrmRunState) LrmRunState {
	contextLog := log.WithFields(log.Fields{"resource": rscName})
	newRunState := runState
	stopEntry := lrmRsc.FindElement(cibTagLrmRscOp + "[@" + cibAttrKeyOperation + "='" + cibAttrValueStop + "']")
	if stopEntry != nil {
		rc, err := getLrmRcCode(rscName, stopEntry)
		if err != nil {
			contextLog.Warning(err)
		} else if rc == ocfSuccess {
			if newRunState == Unknown {
				newRunState = Stopped
			}
		} else {
			newRunState = Running
		}

		return newRunState
	}

	monEntry := lrmRsc.FindElement(cibTagLrmRscOp + "[@" + cibAttrKeyOperation + "='" + cibAttrValueMonitor + "']")
	if monEntry != nil {
		rc, err := getLrmRcCode(rscName, monEntry)
		if err != nil {
			contextLog.Warning(err)
		} else if rc == ocfNotRunning {
			if newRunState == Unknown {
				newRunState = Stopped
			}
		} else {
			newRunState = Running
		}

		return newRunState
	}

	startEntry := lrmRsc.FindElement(cibTagLrmRscOp + "[@" + cibAttrKeyOperation + "='" + cibAttrValueStart + "']")
	if startEntry != nil {
		rc, err := getLrmRcCode(rscName, startEntry)
		if err != nil {
			contextLog.Warning(err)
		} else if rc == ocfRunningMaster || rc == ocfSuccess {
			if newRunState == Unknown {
				newRunState = Running
			}
		} else {
			newRunState = Stopped
		}

		return newRunState
	}

	return newRunState
}

// getLrmRcCode extracts the rc-code value from an LRM operation entry
func getLrmRcCode(rscName string, entry *xmltree.Element) (int, error) {
	rcAttr := entry.SelectAttr(cibAttrKeyRcCode)
	if rcAttr == nil {
		return 0, errors.New("Found LRM resource operation data without a status code")
	}

	rc, err := strconv.ParseInt(rcAttr.Value, 10, 16)
	return int(rc), err
}
