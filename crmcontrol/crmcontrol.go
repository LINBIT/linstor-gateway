package crmcontrol

import "fmt"
import "strings"
import "strconv"
import "errors"
import "github.com/LINBIT/linstor-remote-storage/templateproc"
import "github.com/LINBIT/linstor-remote-storage/extcmd"
import "github.com/LINBIT/linstor-remote-storage/debug"
import xmltree "github.com/beevik/etree"

const (
	CRM_TMPL     = "templates/crm-iscsi.tmpl"
	TGT_LOC_TMPL = "templates/target-location-nodes.tmpl"
	LU_LOC_TMPL  = "templates/lu-location-nodes.tmpl"
)

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

const (
	CIB_RSC_XPATH    = "/cib/configuration/resources"
	CIB_RSC_ATTR     = "instance_attributes"
	CIB_RSC_ATTR_KEY = "nvpair"
)

const (
	CRM_ISCSI_RSC_PREFIX  = "p_iscsi_"
	CRM_ISCSI_LU_NAME     = "lu"
	CRM_ISCSI_PRM_TID     = "tid"
	CRM_TYPE_ISCSI_TARGET = "iSCSITarget"
	CRM_TYPE_ISCSI_LU     = "iSCSILogicalUnit"
)

const (
	CIB_TAG_LOCATION   = "rsc_location"
	CIB_TAG_COLOCATION = "rsc_colocation"
	CIB_TAG_ORDER      = "rsc_order"
	CIB_TAG_RSC_REF    = "resource_ref"
	CIB_TAG_LRM_RSC    = "lrm_resource"
)

const MAX_RECURSION_LEVEL = 40

type CrmConfiguration struct {
	TargetList   []string
	LuList       []string
	OtherRscList []string
	TidSet       TargetIdSet
}

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

	targetLocData, err := constructNodesTemplate(TGT_LOC_TMPL, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	luLocData, err := constructNodesTemplate(LU_LOC_TMPL, storageNodeList, tmplVars)
	if err != nil {
		return err
	}
	tmplVars[VAR_TGT_LOC_NODES] = targetLocData
	tmplVars[VAR_LU_LOC_NODES] = luLocData

	cibData := templateproc.ReplaceVariables(tmplLines, tmplVars)

	// debug.PrintfLnCaption("CIB update input:")
	// debug.PrintTextArray(cibData)

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	cmd, cmdPipe, err := extcmd.PipeToExtCmd("cibadmin", []string{"--modify", "--allow-create", "--xml-pipe"})
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

func DeleteCrmLu(
	iscsiTargetName string,
	lun uint8,
) error {
	luName := CRM_ISCSI_LU_NAME + strconv.Itoa(int(lun))

	crmLu := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_" + luName
	crmTgt := CRM_ISCSI_RSC_PREFIX + iscsiTargetName
	crmSvcIp := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_ip"

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

	err = dissolveConstraints(cib, delItems)
	if err != nil {
		return err
	}

	fmt.Printf("Executing dissolveConstraints(...) again\n")
	err = dissolveConstraints(cib, delItems)
	if err != nil {
		return err
	}

	fmt.Printf("Deleting resources:\n")
	for elemId, _ := range delItems {
		rscElem := cib.FindElement("/cib/configuration/resources/primitive[@id='" + elemId + "']")
		if rscElem != nil {
			rscElemParent := rscElem.Parent()
			if rscElemParent != nil {
				fmt.Printf("Deleting '%s'\n", elemId)
				rscElemParent.RemoveChildAt(rscElem.Index())
			} else {
				return errors.New("Cannot modify CIB, CRM resource '" + elemId + "' has no parent object")
			}
		} else {
			fmt.Printf("Warning: CRM resource '%s' not found in the CIB\n", elemId)
		}
	}

	cibData, err := docRoot.WriteToString()
	if err != nil {
		return err
	}

	// Call cibadmin and pipe the CIB update data to the cluster resource manager
	cmd, cmdPipe, err := extcmd.PipeToExtCmd("cibadmin", []string{"--replace", "--xml-pipe"})
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

func ParseConfiguration(docRoot *xmltree.Document) (CrmConfiguration, error) {
	config := CrmConfiguration{TidSet: NewTargetIdSet()}
	if docRoot == nil {
		return config, errors.New("Internal error: ParseConfiguration() called with docRoot == nil")
	}

	cib := docRoot.Root()
	if cib == nil {
		return config, errors.New("Failed to find the cluster information base (CIB) root element")
	}

	rscSection := cib.FindElement(CIB_RSC_XPATH)
	if rscSection == nil {
		return config, errors.New("Failed to find the cluster resources section in the cluster information base (CIB)")
	}

	resources := rscSection.ChildElements()
	if resources == nil {
		return config, errors.New("Failed to find any cluster resources in the cluster information base (CIB)")
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

	return config, nil
}

func ReadConfiguration() (*xmltree.Document, error) {
	cmd, _, err := extcmd.PipeToExtCmd("cibadmin", []string{"--query"})
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

func copyMap(srcMap map[string]string) map[string]string {
	resultMap := make(map[string]string, len(srcMap))
	for key, value := range srcMap {
		resultMap[key] = value
	}
	return resultMap
}

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

func isTargetEntry(typeAttr xmltree.Attr) bool {
	return typeAttr.Value == CRM_TYPE_ISCSI_TARGET
}

func isLogicalUnitEntry(typeAttr xmltree.Attr) bool {
	return typeAttr.Value == CRM_TYPE_ISCSI_LU
}

func getRscParams(resource *xmltree.Element) []*xmltree.Element {
	var attrList []*xmltree.Element
	instAttr := resource.FindElement(CIB_RSC_ATTR)
	if instAttr != nil {
		attrList = instAttr.ChildElements()
	}
	return attrList
}

type SortUint8 []uint8

func (data SortUint8) Len() int {
	return len(data)
}

func (data SortUint8) Swap(idx1st int, idx2nd int) {
	data[idx1st], data[idx2nd] = data[idx2nd], data[idx1st]
}

func (data SortUint8) Less(idx1st int, idx2nd int) bool {
	return data[idx1st] < data[idx2nd]
}

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

func dissolveConstraints(cibElem *xmltree.Element, delItems map[string]interface{}) error {
	return dissolveConstraintsImpl(cibElem, delItems, 0)
}

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
	if delIdxSet != nil {
		delIdxIter := delIdxSet.Iterator()
		for delIdx, valid := delIdxIter.Next(); valid; delIdx, valid = delIdxIter.Next() {
			cibElem.RemoveChildAt(delIdx)
		}
	}

	return nil
}

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

func isLocationDependency(cibElem *xmltree.Element, delItems map[string]interface{}) bool {
	depFlag := false

	rscAttr := cibElem.SelectAttr("rsc")
	if rscAttr != nil {
		_, depFlag = delItems[rscAttr.Value]
	}

	return depFlag
}

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

func isLrmDependency(cibElem *xmltree.Element, delItems map[string]interface{}) (bool, error) {
	depFlag := false

	idAttr := cibElem.SelectAttr("id")
	if idAttr == nil {
		return false, errors.New("Unparseable " + cibElem.Tag + " constraint, cannot find \"id\" attribute")
	}

	_, depFlag = delItems[idAttr.Value]

	return depFlag, nil
}

func maxRecursionError() error {
	return errors.New("Exceeding maximum recursion level, operation aborted")
}
