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

import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/templateproc"
import "github.com/LINBIT/linstor-remote-storage/extcmd"
import "github.com/LINBIT/linstor-remote-storage/debug"
import xmltree "github.com/beevik/etree"

// Template file names
const (
	CRM_TMPL     = "templates/crm-iscsi.tmpl"
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

// Pacemaker CIB XML resource XPaths and resource sub tag names
// FIXME: the resource sub tag names should probably go into the CIB_TAG_... section
const (
	CIB_RSC_XPATH    = "/cib/configuration/resources"
	CIB_RSC_ATTR     = "instance_attributes"
	CIB_RSC_ATTR_KEY = "nvpair"
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
	CIB_TAG_LRM_RSC    = "lrm_resource"
)

// Maximum recursion level, currently used to limit recursion during recursive
// searches of the XML document tree
const MAX_RECURSION_LEVEL = 40

// Data structure for collecting information about (Pacemaker) CRM resources
type CrmConfiguration struct {
	TargetList   []string
	LuList       []string
	OtherRscList []string
	TidSet       TargetIdSet
}

// Creates the CRM resources
//
// The resources created depend on the contents of the template for resource creation.
// Typically, it's an iSCSI target, logical unit and service IP address, along
// with constraints that bundle them and place them on the selected nodes
func CreateCrmLu(
	storageNodeList []string,
	iscsiTargetName string,
	ip string,
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
	tmplVars[VAR_SVC_IP] = ip
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

	fmt.Printf("CRM command execution successful\n\n")

	if len(stdoutLines) >= 1 {
		fmt.Printf("\x1b[1;33mBegin of CRM command stdout output:\x1b[0m\n")
		debug.PrintTextArrayLimited(stdoutLines, 5)
	} else {
		fmt.Printf("No stdout output\n")
	}

	if len(stderrLines) >= 1 {
		fmt.Printf("\x1b[1;33mCRM command stderr output:\x1b[0m\n")
		fmt.Printf("\x1b[1;31m")
		debug.PrintTextArray(stderrLines)
		fmt.Printf("\x1b[0m")
	} else {
		fmt.Printf("No stderr output\n")
	}

	return err
}

// Deletes the CRM resources
//
// TODO:
// The names of the objects to delete are currently hard coded, they should probably
// be loaded from a file, because the actual objects created depend on templates that
// are also stored in files.
func DeleteCrmLu(
	iscsiTargetName string,
	lun uint8,
) error {
	luName := CRM_ISCSI_LU_NAME + strconv.Itoa(int(lun))

	crmLu := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_" + luName
	crmTgt := CRM_ISCSI_RSC_PREFIX + iscsiTargetName
	crmSvcIp := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_ip"

	// Read the current CIB XML
	docRoot, err := ReadConfiguration()
	if err != nil {
		return err
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	delItems := make(map[string]interface{})
	delItems[crmTgt] = nil
	delItems[crmLu] = nil
	delItems[crmSvcIp] = nil

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

	// Serialize the modified XML document tree into a string containing the XML document (CIB update data)
	cibData, err := docRoot.WriteToString()
	if err != nil {
		return err
	}

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	cmd, cmdPipe, err := extcmd.PipeToExtCmd(CRM_DELETE_COMMAND.executable, CRM_DELETE_COMMAND.arguments)
	if err != nil {
		return err
	}

	fmt.Print("Updated CIB data:\n")
	fmt.Print(cibData)
	fmt.Print("\n\n")

	_, err = cmdPipe.WriteString(cibData)
	if err != nil {
		cmd.IoFailed()
	}
	cmdPipe.Flush()

	stdoutLines, stderrLines, err := cmd.WaitForExtCmd()

	fmt.Printf("CRM command execution successful\n\n")

	if len(stdoutLines) >= 1 {
		fmt.Printf("\x1b[1;33mBegin of CRM command stdout output:\x1b[0m\n")
		debug.PrintTextArrayLimited(stdoutLines, 5)
	} else {
		fmt.Printf("No stdout output\n")
	}

	if len(stderrLines) >= 1 {
		fmt.Printf("\x1b[1;33mCRM command stderr output:\x1b[0m\n")
		fmt.Printf("\x1b[1;31m")
		debug.PrintTextArray(stderrLines)
		fmt.Printf("\x1b[0m")
	} else {
		fmt.Printf("No stderr output\n")
	}

	return err
}

// Parses the CIB XML document and returns information about existing resources
//
// Information about existing CRM resources is parsed from the CIB XML document and
// stored in a newly allocated CrmConfiguration data structure
func ParseConfiguration(docRoot *xmltree.Document) (*CrmConfiguration, error) {
	config := CrmConfiguration{TidSet: NewTargetIdSet()}
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
	if resources == nil {
		return nil, errors.New("Failed to find any cluster resources in the cluster information base (CIB)")
	}

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
		fmt.Printf("\x1b[1;33m")
		fmt.Printf("External command error output:")
		fmt.Printf("\x1b[0m\n")
		debug.PrintTextArray(stderrLines)
		fmt.Printf("\n")
	}

	docData := extcmd.FuseStrings(stdoutLines)
	docRoot := xmltree.NewDocument()
	err = docRoot.ReadFromString(docData)
	if err != nil {
		return nil, err
	}

	return docRoot, nil
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
	instAttr := resource.FindElement(CIB_RSC_ATTR)
	if instAttr != nil {
		attrList = instAttr.ChildElements()
	}
	return attrList
}

// Prints collected stdout/stderr output of an external command, or indicates
// that the external command did not produce such output
func printCmdOutput(stdoutLines []string, stderrLines []string) {
	if len(stdoutLines) > 0 {
		fmt.Printf("Stdout output:\n")
		debug.PrintTextArray(stdoutLines)
	} else {
		fmt.Printf("No stdout output\n")
	}

	if len(stderrLines) > 0 {
		fmt.Printf("Stderr output:\n")
		debug.PrintTextArray(stderrLines)
	} else {
		fmt.Printf("No stderr output\n")
	}

	fmt.Printf("\n")
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

// Generates an error indicating that an operation was aborted because it reached the maximum recursion level
func maxRecursionError() error {
	return errors.New("Exceeding maximum recursion level, operation aborted")
}
