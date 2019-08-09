// CRM (Pacemaker) API
package crmcontrol

// crmcontrol module
//
// The functions in this module are called by the high-level API in package application
// (module application.go) to perform operations in the CRM subsystem, such as
// creating the primitives and constraints that configure iSCSI targets, logical units
// and the associated service IP addresses.
// The 'cibadmin' utility is used to modify the cluster's CIB (cluster information base).
// The CIB is modified by
//   - sending XML entries, created from templates, to create new primitives & constraints,
//     much like a macro processor
//   - reading and parsing the current CIB XML, and modifying the contents
//     (e.g. removing tags and their nested tags) to delete existing entries from
//     the cluster configuration.
// The 'etree' package is used for XML parsing and modification.

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/LINBIT/linstor-remote-storage/extcmd"
	"github.com/LINBIT/linstor-remote-storage/templateproc"
	log "github.com/sirupsen/logrus"

	xmltree "github.com/beevik/etree"
)

// Template file names
const (
	CRM_TMPL     = "templates/crm-iscsi.tmpl"
	CRM_OBJ_TMPL = "templates/crm-obj-names.tmpl"
	TGT_LOC_TMPL = "templates/target-location-nodes.tmpl"
	LU_LOC_TMPL  = "templates/lu-location-nodes.tmpl"
)

// Template variable keys
const (
	VAR_NODE_NAME     = "CRM_NODE_NAME"
	VAR_NR            = "NR"
	VAR_LU_NAME       = "CRM_LU_NAME"
	VAR_SVC_IP        = "CRM_SVC_IP"
	VAR_TGT_NAME      = "CRM_TARGET_NAME"
	VAR_TGT_IQN       = "TARGET_IQN"
	VAR_ISCSI_LUN     = "LUN"
	VAR_STOR_DEV      = "DEVICE"
	VAR_USERNAME      = "USERNAME"
	VAR_PASSWORD      = "PASSWORD"
	VAR_PORTALS       = "PORTALS"
	VAR_TID           = "TID"
	VAR_TGT_LOC_NODES = "TARGET_LOCATION_NODES"
	VAR_LU_LOC_NODES  = "LU_LOCATION_NODES"
)

// Pacemaker CIB XML XPaths
const (
	CIB_RSC_XPATH    = "/cib/configuration/resources"
	CIB_STATUS_XPATH = "/cib/status"
)

// Pacemaker CRM resource names, prefixes, suffixes, search patterns, etc.
const (
	CRM_ISCSI_RSC_PREFIX  = "p_iscsi_"
	CRM_ISCSI_LU_NAME     = "lu"
	CRM_ISCSI_PRM_TID     = "tid"
	CRM_TYPE_ISCSI_TARGET = "iSCSITarget"
	CRM_TYPE_ISCSI_LU     = "iSCSILogicalUnit"
)

// Pacemaker CIB XML tag names
const (
	CIB_TAG_LOCATION   = "rsc_location"
	CIB_TAG_COLOCATION = "rsc_colocation"
	CIB_TAG_ORDER      = "rsc_order"
	CIB_TAG_RSC_REF    = "resource_ref"
	CIB_TAG_META_ATTR  = "meta_attributes"
	CIB_TAG_INST_ATTR  = "instance_attributes"
	CIB_TAG_NV_PAIR    = "nvpair"
	CIB_TAG_LRM        = "lrm"
	CIB_TAG_LRM_RSCLST = "lrm_resources"
	CIB_TAG_LRM_RSC    = "lrm_resource"
	CIB_TAG_LRM_RSC_OP = "lrm_rsc_op"
)

// Pacemaker CIB attribute names
const (
	CIB_ATTR_KEY_ID            = "id"
	CIB_ATTR_KEY_NAME          = "name"
	CIB_ATTR_KEY_VALUE         = "value"
	CIB_ATTR_KEY_OPERATION     = "operation"
	CIB_ATTR_KEY_RC_CODE       = "rc-code"
	CIB_ATTR_VALUE_TARGET_ROLE = "target-role"
	CIB_ATTR_VALUE_STARTED     = "Started"
	CIB_ATTR_VALUE_STOPPED     = "Stopped"
	CIB_ATTR_VALUE_STOP        = "stop"
	CIB_ATTR_VALUE_MONITOR     = "monitor"
)

// Pacemaker OCF resource agent exit codes
const (
	OCF_SUCCESS           = 0
	OCF_ERR_GENERIC       = 1
	OCF_ERR_ARGS          = 2
	OCF_ERR_UNIMPLEMENTED = 3
	OCF_ERR_PERM          = 4
	OCF_ERR_INSTALLED     = 5
	OCF_NOT_RUNNING       = 7
	OCF_RUNNING_MASTER    = 8
	OCF_FAILED_MASTER     = 9
)

// Maximum recursion level, currently used to limit recursion during recursive
// searches of the XML document tree
const MAX_RECURSION_LEVEL = 40

// Maximum number of CIB poll retries when waiting for CRM resources to stop
const MAX_WAIT_STOP_RETRIES = 10

// Initial delay after setting resource target-role=Stopped before starting to poll the CIB
// to check whether resources have actually stopped
const WAIT_STOP_POLL_CIB_DELAY = 2500

// Delay between CIB polls in milliseconds
const CIB_POLL_RETRY_DELAY = 2000

// Data structure for collecting information about (Pacemaker) CRM resources
type CrmConfiguration struct {
	RscMap       map[string]interface{}
	TargetList   []string
	LuList       []string
	OtherRscList []string
	TidSet       TargetIdSet
}

type LrmRunState struct {
	HaveState bool
	Running   bool
}

// Creates the CRM resources
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
	tmplLines, err := templateproc.LoadTemplate(CRM_TMPL)
	if err != nil {
		return err
	}

	// debug.PrintfLnCaption("Template input:")
	// debug.PrintTextArray(tmplLines)

	// Construct the CIB update data from the template
	var tmplVars map[string]string = make(map[string]string)
	tmplVars[VAR_LU_NAME] = "lu" + strconv.Itoa(int(lun))
	tmplVars[VAR_SVC_IP] = ip.String()
	tmplVars[VAR_TGT_NAME] = iscsiTargetName
	tmplVars[VAR_TGT_IQN] = iscsiTargetIqn
	tmplVars[VAR_ISCSI_LUN] = strconv.Itoa(int(lun))
	tmplVars[VAR_STOR_DEV] = device
	tmplVars[VAR_USERNAME] = username
	tmplVars[VAR_PASSWORD] = password
	tmplVars[VAR_PORTALS] = portal
	tmplVars[VAR_TID] = strconv.Itoa(int(tid))

	// Create sub XML content, one entry per node, from the iSCSI target location constraint template
	targetLocData, err := constructNodesTemplate(TGT_LOC_TMPL, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	// Create sub XML content, one entry per node, from the iSCSI logical unit location constraint template
	luLocData, err := constructNodesTemplate(LU_LOC_TMPL, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	// Load the sub XML content into variables
	tmplVars[VAR_TGT_LOC_NODES] = targetLocData
	tmplVars[VAR_LU_LOC_NODES] = luLocData

	// Replace resource creation template variables
	cibData := templateproc.ReplaceVariables(tmplLines, tmplVars)

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	cmd, cmdPipe, err := extcmd.PipeToExtCmd(CRM_CREATE_COMMAND.executable, CRM_CREATE_COMMAND.arguments)
	if err != nil {
		return err
	}

	for _, line := range cibData {
		_, err := cmdPipe.WriteString(line)
		if err != nil {
			cmd.IoFailed()
			break
		}
	}
	cmdPipe.Flush()

	stdoutLines, stderrLines, err := cmd.WaitForExtCmd()

	log.Info("CRM command execution successful")

	if len(stdoutLines) >= 1 {
		log.Debug("Begin of CRM command stdout output:", stdoutLines)
	} else {
		log.Debug("No stdout output")
	}

	if len(stderrLines) >= 1 {
		log.Debug("CRM command stderr output:", stderrLines)
	} else {
		log.Debug("No stdout output")
	}

	return err
}

// Stops the CRM resources
func ModifyCrmLuTargetRole(
	iscsiTargetName string,
	lun uint8,
	startFlag bool,
) error {
	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return err
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	stopItems, err := LoadCrmObjMap(iscsiTargetName, lun)
	if err != nil {
		return err
	}

	// Process the CIB XML document tree and insert meta attributes for target-role=Stopped
	for elemId, _ := range stopItems {
		rscElem := cib.FindElement("/cib/configuration/resources/primitive[@id='" + elemId + "']")
		if rscElem != nil {
			var tgtRoleEntry *xmltree.Element = nil
			metaAttr := rscElem.FindElement(CIB_TAG_META_ATTR)
			if metaAttr != nil {
				// Meta attributes exist, find the entry that sets the target-role
				tgtRoleEntry = metaAttr.FindElement(CIB_TAG_NV_PAIR + "[@" + CIB_ATTR_KEY_NAME + "='" + CIB_ATTR_VALUE_TARGET_ROLE + "']")
			} else {
				// No meta attributes present, create XML element
				metaAttr = rscElem.CreateElement(CIB_TAG_META_ATTR)
				metaAttr.CreateAttr(CIB_ATTR_KEY_ID, elemId+"_"+CIB_TAG_META_ATTR)
			}
			if tgtRoleEntry == nil {
				// No entry that sets the target-role, create entry
				tgtRoleEntry = metaAttr.CreateElement(CIB_TAG_NV_PAIR)
				tgtRoleEntry.CreateAttr(CIB_ATTR_KEY_ID, elemId+"_"+CIB_ATTR_VALUE_TARGET_ROLE)
				tgtRoleEntry.CreateAttr(CIB_ATTR_KEY_NAME, CIB_ATTR_VALUE_TARGET_ROLE)
			}
			// Set the target-role
			var tgtRoleValue string
			if startFlag {
				tgtRoleValue = CIB_ATTR_VALUE_STARTED
			} else {
				tgtRoleValue = CIB_ATTR_VALUE_STOPPED
			}
			tgtRoleEntry.CreateAttr(CIB_ATTR_KEY_VALUE, tgtRoleValue)
		} else {
			fmt.Printf("Warning: CRM resource '%s' not found in the CIB\n", elemId)
		}
	}

	return executeCibUpdate(docRoot, CRM_UPDATE_COMMAND)
}

// Deletes the CRM resources
func DeleteCrmLu(
	iscsiTargetName string,
	lun uint8,
) error {
	err := ModifyCrmLuTargetRole(iscsiTargetName, lun, false)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(WAIT_STOP_POLL_CIB_DELAY * time.Millisecond))
	isStopped, err := WaitForResourceStop(iscsiTargetName, lun)
	if err != nil {
		return err
	}

	if !isStopped {
		return errors.New("Resource stop was not confirmed for all resources, cannot continue delete action")
	}

	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return err
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	delItems, err := LoadCrmObjMap(iscsiTargetName, lun)
	if err != nil {
		return err
	}

	// Process the CIB XML document tree, removing constraints that refer to any of the objects
	// that will be deleted
	err = dissolveConstraints(cib, delItems)
	if err != nil {
		return err
	}

	// Process the CIB XML document tree, removing the specified CRM resources
	for elemId, _ := range delItems {
		rscElem := cib.FindElement("/cib/configuration/resources/primitive[@id='" + elemId + "']")
		if rscElem != nil {
			rscElemParent := rscElem.Parent()
			if rscElemParent != nil {
				rscElemParent.RemoveChildAt(rscElem.Index())
			} else {
				return errors.New("Cannot modify CIB, CRM resource '" + elemId + "' has no parent object")
			}
		} else {
			fmt.Printf("Warning: CRM resource '%s' not found in the CIB\n", elemId)
		}
	}

	return executeCibUpdate(docRoot, CRM_UPDATE_COMMAND)
}

// Probes the LRM run state of the CRM resources associated with the specified iSCSI resource
func ProbeResource(
	iscsiTargetName string,
	lun uint8,
) (map[string]LrmRunState, error) {
	rscStateMap := make(map[string]LrmRunState)

	stopItems, err := LoadCrmObjMap(iscsiTargetName, lun)
	if err != nil {
		return rscStateMap, err
	}

	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return rscStateMap, err
	}

	err = probeResourceRunState(&stopItems, docRoot)
	if err != nil {
		return rscStateMap, err
	}

	for rscName, tmpRunState := range stopItems {
		runState := tmpRunState.(LrmRunState)
		rscStateMap[rscName] = runState
	}

	return rscStateMap, nil
}

// Waits for CRM resources to stop
//
// Returns: Flag indicating whether resources are stopped (true) or not (false), error object
func WaitForResourceStop(
	iscsiTargetName string,
	lun uint8,
) (bool, error) {
	stopItems, err := LoadCrmObjMap(iscsiTargetName, lun)
	if err != nil {
		return false, err
	}

	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return false, err
	}

	config, err := ParseConfiguration(docRoot)
	if err != nil {
		return false, err
	}

	for rscName, _ := range stopItems {
		_, found := config.RscMap[rscName]
		if !found {
			fmt.Printf("Warning: Resource '%s' not found in the CIB\n    This resource will be ignored.\n", rscName)
		}
		delete(stopItems, rscName)
	}

	fmt.Print("Waiting for the following CRM resources to stop:\n")
	for rscName, _ := range stopItems {
		fmt.Printf("    %s\n", rscName)
	}

	isStopped := false
	retries := 0
	for !isStopped {
		err := probeResourceRunState(&stopItems, docRoot)
		if err != nil {
			return false, err
		}

		_, stoppedFlag := checkResourceStopped(&stopItems)

		if !stoppedFlag {
			if retries > MAX_WAIT_STOP_RETRIES {
				break
			}

			time.Sleep(time.Duration(CIB_POLL_RETRY_DELAY * time.Millisecond))

			// Re-read the current CIB XML
			docRoot, err = ReadConfiguration()
			if err != nil {
				return false, err
			}
		} else {
			isStopped = true
		}

		retries++
	}

	if isStopped {
		fmt.Printf("The resources are stopped")
	} else {
		fmt.Printf("Could not confirm that the resources are stopped")
	}

	return isStopped, nil
}

// Parses the CIB XML document and returns information about existing resources
//
// Information about existing CRM resources is parsed from the CIB XML document and
// stored in a newly allocated CrmConfiguration data structure
func ParseConfiguration(docRoot *xmltree.Document) (*CrmConfiguration, error) {
	config := CrmConfiguration{RscMap: make(map[string]interface{}), TidSet: NewTargetIdSet()}
	if docRoot == nil {
		return nil, errors.New("Internal error: ParseConfiguration() called with docRoot == nil")
	}

	cib := docRoot.Root()
	if cib == nil {
		return nil, errors.New("Failed to find the cluster information base (CIB) root element")
	}

	rscSection := cib.FindElement(CIB_RSC_XPATH)
	if rscSection == nil {
		return nil, errors.New("Failed to find the cluster resources section in the cluster information base (CIB)")
	}

	resources := rscSection.ChildElements()

	for _, selectedRsc := range resources {
		idAttr := selectedRsc.SelectAttr("id")
		if idAttr != nil {
			crmRscName := idAttr.Value
			isTarget := false
			isLu := false
			typeAttr := selectedRsc.SelectAttr("type")
			if typeAttr != nil {
				isTarget = isTargetEntry(*typeAttr)
				isLu = isLogicalUnitEntry(*typeAttr)
			} else {
				fmt.Printf("Warning: CIB primitive element has no attribute \x1b[1;32mtype\x1b[0m\n")
			}

			config.RscMap[crmRscName] = nil
			if isTarget {
				config.TargetList = append(config.TargetList, crmRscName)
			} else if isLu {
				config.LuList = append(config.LuList, crmRscName)
			} else {
				config.OtherRscList = append(config.OtherRscList, crmRscName)
			}

			tidEntry := selectedRsc.FindElement("instance_attributes/nvpair[@name='tid']")
			if tidEntry != nil {
				tidAttr := tidEntry.SelectAttr("value")
				if tidAttr != nil {
					tid, err := strconv.ParseInt(tidAttr.Value, 10, 16)
					if err != nil {
						fmt.Printf("\x1b[1;31mWarning: Unparseable tid parameter '%s' for resource '%s'\x1b[0m\n", tidAttr.Value, idAttr.Value)
					}
					if tid > 0 {
						config.TidSet.Insert(int16(tid))
					} else {
						fmt.Printf("\x1b[1;31mWarning: Invalid tid value %d for resource '%s'\x1b[0m\n", tid, idAttr.Value)
					}

				}
			}
		} else {
			fmt.Printf("Warning: CIB primitive element has no attribute \x1b[1;32mname\x1b[0m\n")
		}
	}

	return &config, nil
}

// Reads the CIB XML document into a string
func ReadConfiguration() (*xmltree.Document, error) {
	cmd, _, err := extcmd.PipeToExtCmd(CRM_LIST_COMMAND.executable, CRM_LIST_COMMAND.arguments)
	if err != nil {
		return nil, err
	}
	stdoutLines, stderrLines, err := cmd.WaitForExtCmd()
	if len(stderrLines) > 0 {
		log.Debug("External command error output:", stderrLines)
	}

	docData := extcmd.FuseStrings(stdoutLines)
	docRoot := xmltree.NewDocument()
	err = docRoot.ReadFromString(docData)
	if err != nil {
		return nil, err
	}

	return docRoot, nil
}

// Loads a map of CRM object names from the template
func LoadCrmObjMap(iscsiTargetName string, lun uint8) (map[string]interface{}, error) {
	objMap := make(map[string]interface{})
	nameTmplList, err := templateproc.LoadTemplate(CRM_OBJ_TMPL)
	if err != nil {
		return objMap, err
	}
	tmplVars := make(map[string]string)
	tmplVars[VAR_TGT_NAME] = iscsiTargetName
	tmplVars[VAR_LU_NAME] = CRM_ISCSI_LU_NAME + strconv.Itoa(int(lun))
	nameList := templateproc.ReplaceVariables(nameTmplList, tmplVars)
	for _, nameLine := range nameList {
		name := strings.TrimRight(nameLine, "\r\n")
		objMap[name] = nil
	}
	return objMap, nil
}

// Probes the LRM run state of selected resources
//
// Each resource name is mapped to an LrmRunState data structure that is then
// updated with the run state of the respective resource
func probeResourceRunState(stopItems *map[string]interface{}, docRoot *xmltree.Document) error {
	for key, _ := range *stopItems {
		(*stopItems)[key] = LrmRunState{}
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	statusSection := cib.FindElement(CIB_STATUS_XPATH)
	if statusSection == nil {
		return errors.New("Failed to find any resource status information in the cluster information base (CIB)")
	}

	for _, nodeElem := range statusSection.ChildElements() {
		lrmElem := nodeElem.SelectElement(CIB_TAG_LRM)
		if lrmElem != nil {
			lrmRscList := lrmElem.SelectElement(CIB_TAG_LRM_RSCLST)
			if lrmRscList != nil {
				for _, lrmRsc := range lrmRscList.ChildElements() {
					idAttr := lrmRsc.SelectAttr(CIB_ATTR_KEY_ID)
					if idAttr == nil {
						return errors.New("Unparseable " + lrmRsc.Tag + " entry, cannot find \"id\" attribute")
					}
					rscName := idAttr.Value
					tmpRunState, isStopItem := (*stopItems)[rscName]
					if isStopItem {
						itemRunState := tmpRunState.(LrmRunState)
						updateRunState(rscName, lrmRsc, &itemRunState)
						(*stopItems)[rscName] = itemRunState
					}
				}
			}
		}
	}

	return nil
}

func checkResourceStopped(stopItems *map[string]interface{}) (bool, bool) {
	stateCtr := 0
	stoppedCtr := 0
	for rscName, tmpRunState := range *stopItems {
		runState := tmpRunState.(LrmRunState)
		if runState.HaveState {
			stateCtr++
			if runState.Running {
				fmt.Printf("Resource '%s' is running\n", rscName)
			} else {
				fmt.Printf("Resource '%s' is stopped\n", rscName)
				stoppedCtr++
			}
		} else {
			fmt.Printf("Warning: No run status information for resource '%s'\n", rscName)
		}
	}

	haveState := stateCtr == len(*stopItems)
	stoppedFlag := stoppedCtr == len(*stopItems)

	return haveState, stoppedFlag
}

func executeCibUpdate(docRoot *xmltree.Document, crmCmd CrmCommand) error {
	// Serialize the modified XML document tree into a string containing the XML document (CIB update data)
	cibData, err := docRoot.WriteToString()
	if err != nil {
		return err
	}

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	cmd, cmdPipe, err := extcmd.PipeToExtCmd(crmCmd.executable, crmCmd.arguments)
	if err != nil {
		return err
	}

	_, err = cmdPipe.WriteString(cibData)
	if err != nil {
		cmd.IoFailed()
	}
	cmdPipe.Flush()

	stdoutLines, stderrLines, err := cmd.WaitForExtCmd()

	if err == nil {
		fmt.Print("CRM command execution successful\n\n")
	} else {
		fmt.Print("CRM command execution returned an error\n\n")
		fmt.Print("The updated CIB data sent to the command was:\n")
		fmt.Print(cibData)
		fmt.Print("\n\n")
	}

	if len(stdoutLines) >= 1 {
		log.Debug("Begin of CRM command stdout output:", stdoutLines)
	} else {
		log.Debug("No stdout output\n")
	}

	if len(stderrLines) >= 1 {
		log.Debug("CRM command stderr output:", stderrLines)
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
func constructNodesTemplate(srcFile string, nodeList []string, tmplVars map[string]string) (string, error) {
	subTmplLines, err := templateproc.LoadTemplate(srcFile)
	if err != nil {
		return "", err
	}
	subTmplVars := copyMap(tmplVars)
	var nr uint32 = 0
	var subDataBld strings.Builder
	for _, nodename := range nodeList {
		subTmplVars[VAR_NODE_NAME] = nodename
		subTmplVars[VAR_NR] = strconv.FormatUint(uint64(nr), 10)
		subDataLines := templateproc.ReplaceVariables(subTmplLines, subTmplVars)
		for _, line := range subDataLines {
			subDataBld.WriteString(line)
		}
		nr++
	}
	return subDataBld.String(), nil
}

// Identifies CRM iSCSI target resources by checking the resource agent name
func isTargetEntry(typeAttr xmltree.Attr) bool {
	return typeAttr.Value == CRM_TYPE_ISCSI_TARGET
}

// Identifies CRM iSCSI logical unit resources by checking the resource agent name
func isLogicalUnitEntry(typeAttr xmltree.Attr) bool {
	return typeAttr.Value == CRM_TYPE_ISCSI_LU
}

// Returns resource attributes, if present, otherwise nil
func getRscParams(resource *xmltree.Element) []*xmltree.Element {
	var attrList []*xmltree.Element
	instAttr := resource.FindElement(CIB_TAG_INST_ATTR)
	if instAttr != nil {
		attrList = instAttr.ChildElements()
	}
	return attrList
}

// Prints collected stdout/stderr output of an external command, or indicates
// that the external command did not produce such output
func printCmdOutput(stdoutLines []string, stderrLines []string) {
	if len(stdoutLines) > 0 {
		log.Debug("Stdout output:", stdoutLines)
	} else {
		log.Debug("No stdout output")
	}

	if len(stderrLines) > 0 {
		log.Debug("Stderr output:", stderrLines)
	} else {
		log.Debug("No stderr output")
	}
}

// Removes CRM constraints that refer to the specified delItems names from the CIB XML document tree
func dissolveConstraints(cibElem *xmltree.Element, delItems map[string]interface{}) error {
	return dissolveConstraintsImpl(cibElem, delItems, 0)
}

// See dissolveConstraints(...)
func dissolveConstraintsImpl(cibElem *xmltree.Element, delItems map[string]interface{}, recursionLevel int) error {
	// delIdxSet is allocated on-demand only if it is required
	var delIdxSet *ElemIdxSet = nil

	childList := cibElem.ChildElements()
	for _, subElem := range childList {
		var dependFlag bool = false
		var err error
		if subElem.Tag == CIB_TAG_COLOCATION {
			if recursionLevel < MAX_RECURSION_LEVEL {
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
				return maxRecursionError()
			}
		} else if subElem.Tag == CIB_TAG_ORDER {
			if recursionLevel < MAX_RECURSION_LEVEL {
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
				return maxRecursionError()
			}
		} else if subElem.Tag == CIB_TAG_LOCATION {
			if recursionLevel < MAX_RECURSION_LEVEL {
				dependFlag = isLocationDependency(subElem, delItems)
				if !dependFlag {
					dependFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
					if err != nil {
						return err
					}
				}
			} else {
				return maxRecursionError()
			}
		} else if subElem.Tag == CIB_TAG_LRM_RSC {
			if recursionLevel < MAX_RECURSION_LEVEL {
				dependFlag, err = isLrmDependency(subElem, delItems)
				if err != nil {
					return err
				}
			} else {
				return maxRecursionError()
			}
		} else {
			if recursionLevel < MAX_RECURSION_LEVEL {
				err := dissolveConstraintsImpl(subElem, delItems, recursionLevel+1)
				if err != nil {
					return err
				}
			} else {
				return maxRecursionError()
			}
		}
		if dependFlag {
			if delIdxSet == nil {
				setInstance := NewElemIdxSet()
				delIdxSet = &setInstance
			}
			delIdxSet.Insert(subElem.Index())
			idAttr := subElem.SelectAttr("id")
			if idAttr != nil {
				fmt.Printf("Deleting type %s dependency '%s'\n", subElem.Tag, idAttr.Value)
			}
		}
	}
	// Elements are deleted in order of descending index, so that the index of elements
	// deleted later does not change due to reordering elements that had a greater index
	// than an element thas was deleted from the slice/array.
	if delIdxSet != nil {
		delIdxIter := delIdxSet.Iterator()
		for delIdx, valid := delIdxIter.Next(); valid; delIdx, valid = delIdxIter.Next() {
			cibElem.RemoveChildAt(delIdx)
		}
	}

	return nil
}

// Indicates whether an element has sub elements that are resource reference tags that refer to any of the specified delItems names
func hasRscRefDependency(cibElem *xmltree.Element, delItems map[string]interface{}, recursionLevel int) (bool, error) {
	depFlag := false

	var err error
	childList := cibElem.ChildElements()
	for _, subElem := range childList {
		if subElem.Tag == CIB_TAG_RSC_REF {
			idAttr := subElem.SelectAttr("id")
			if idAttr != nil {
				_, depFlag = delItems[idAttr.Value]
			} else {
				return false, errors.New("Unparseable " + subElem.Tag + " tag, cannot find \"id\" attribute")
			}
		} else {
			if recursionLevel < MAX_RECURSION_LEVEL {
				depFlag, err = hasRscRefDependency(subElem, delItems, recursionLevel+1)
				if err != nil {
					return false, err
				}
			} else {
				return false, maxRecursionError()
			}
		}
		if depFlag {
			break
		}
	}

	return depFlag, nil
}

// Indicates whether the element is a CRM location constraint that refers to any of the specified delItems names
func isLocationDependency(cibElem *xmltree.Element, delItems map[string]interface{}) bool {
	depFlag := false

	rscAttr := cibElem.SelectAttr("rsc")
	if rscAttr != nil {
		_, depFlag = delItems[rscAttr.Value]
	}

	return depFlag
}

// Indicates whether the element is a CRM order constraint that refers to any of the specified delItems names
func isOrderDependency(cibElem *xmltree.Element, delItems map[string]interface{}) (bool, error) {
	depFlag := false

	firstAttr := cibElem.SelectAttr("first")
	thenAttr := cibElem.SelectAttr("then")
	if firstAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"first\" attribute")
	}
	if thenAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"then\" attribute")
	}

	_, depFlag = delItems[firstAttr.Value]
	if !depFlag {
		_, depFlag = delItems[thenAttr.Value]
	}

	return depFlag, nil
}

// Indicates whether the element is a CRM colocation constraint that refers to any of the specified delItems names
func isColocationDependency(cibElem *xmltree.Element, delItems map[string]interface{}) (bool, error) {
	depFlag := false

	rscAttr := cibElem.SelectAttr("rsc")
	withRscAttr := cibElem.SelectAttr("with-rsc")
	if rscAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"rsc\" attribute")
	}
	if withRscAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"with-rsc\" attribute")
	}

	_, depFlag = delItems[rscAttr.Value]
	if !depFlag {
		_, depFlag = delItems[withRscAttr.Value]
	}

	return depFlag, nil
}

// Indicates whether the element is an LRM entry that refers to any of the specified delItems names
func isLrmDependency(cibElem *xmltree.Element, delItems map[string]interface{}) (bool, error) {
	depFlag := false

	idAttr := cibElem.SelectAttr("id")
	if idAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"id\" attribute")
	}

	_, depFlag = delItems[idAttr.Value]

	return depFlag, nil
}

// Updates the run status information in the stopItem map
func updateRunState(rscName string, lrmRsc *xmltree.Element, runState *LrmRunState) {
	// For a resource to be considered stopped, this function must find
	// - either a successful stop action
	// - or a monitor action with rc-code OCF_NOT_RUNNING and no stop action
	//
	// If a stop action is present, the monitor action can still show "running" (rc-code OCF_SUCCESS == 0)
	// although the resource is actually stopped. The monitor action's rc-code is only interesting if
	// there is no stop action present.
	stopEntry := lrmRsc.FindElement(CIB_TAG_LRM_RSC_OP + "[@" + CIB_ATTR_KEY_OPERATION + "='" + CIB_ATTR_VALUE_STOP + "']")
	if stopEntry != nil {
		rc, valid := getLrmRcCode(rscName, stopEntry)
		if valid {
			if rc == OCF_SUCCESS {
				if !runState.HaveState {
					runState.HaveState = true
					runState.Running = false
				}
			} else {
				runState.HaveState = true
				runState.Running = true
			}
		}
	} else {
		monEntry := lrmRsc.FindElement(CIB_TAG_LRM_RSC_OP + "[@" + CIB_ATTR_KEY_OPERATION + "='" + CIB_ATTR_VALUE_MONITOR + "']")
		if monEntry != nil {
			rc, valid := getLrmRcCode(rscName, monEntry)
			if valid {
				if rc == OCF_NOT_RUNNING {
					if !runState.HaveState {
						runState.HaveState = true
						runState.Running = false
					}
				} else {
					runState.HaveState = true
					runState.Running = true
				}
			}
		} else {
			startEntry := lrmRsc.FindElement(CIB_TAG_LRM_RSC_OP + "[@" + CIB_ATTR_KEY_OPERATION + "='" + CIB_ATTR_VALUE_MONITOR + "']")
			if startEntry != nil {
				rc, valid := getLrmRcCode(rscName, startEntry)
				if valid {
					switch rc {
					case OCF_RUNNING_MASTER:
						fallthrough
					case OCF_SUCCESS:
						runState.HaveState = true
						runState.Running = true
					default:
						if !runState.HaveState {
							runState.HaveState = true
							runState.Running = false
						}
					}
				}
			}
		}
	}
}

// Extracts the rc-code value from an LRM operation entry
//
// Returns: rc-code value and a flag indicating whether the rc-code is valid (true) or invalid/unusable (false)
func getLrmRcCode(rscName string, entry *xmltree.Element) (int, bool) {
	validFlag := false
	rc := 0
	rcAttr := entry.SelectAttr(CIB_ATTR_KEY_RC_CODE)
	if rcAttr != nil {
		parsedRc, err := strconv.ParseInt(rcAttr.Value, 10, 16)
		if err == nil {
			validFlag = true
			rc = int(parsedRc)
		} else {
			fmt.Printf("Warning: Unparseable LRM resource operation return code for resource '%s'\n", rscName)
		}
	} else {
		fmt.Printf("Warning: Found LRM resource operation data for resource '%s' without a status code\n", rscName)
	}
	return rc, validFlag
}

// Generates an error indicating that an operation was aborted because it reached the maximum recursion level
func maxRecursionError() error {
	return errors.New("Exceeding maximum recursion level, operation aborted")
}
