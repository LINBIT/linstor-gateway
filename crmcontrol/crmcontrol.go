package crmcontrol

import "fmt"
import "strings"
import "strconv"
import "sort"
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
	CRM_ISCSI_RSC_PREFIX = "p_iscsi_"
	CRM_ISCSI_LU_NAME    = "lu"
	CRM_ISCSI_PRM_TID    = "tid"
)

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
	tid string,
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
	tmplVars[VAR_TID] = tid

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

	if err != nil {
		return err
	}

	return nil
}

func ReadConfiguration() error {
	docRoot := xmltree.NewDocument()

	cmd, _, err := extcmd.PipeToExtCmd("cibadmin", []string{"--query"})
	if err != nil {
		return err
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
	err = docRoot.ReadFromString(docData)
	if err != nil {
		return err
	}

	cib := docRoot.Root()
	if cib == nil {
		return errors.New("Failed to find the cluster information base (CIB) root element")
	}

	rscSection := cib.FindElement(CIB_RSC_XPATH)
	if rscSection == nil {
		return errors.New("Failed to find the cluster resources section in the cluster information base (CIB)")
	}

	resources := rscSection.ChildElements()
	if resources == nil {
		return errors.New("Failed to find any cluster resources in the cluster information base (CIB)")
	}

	var tidTable = make(map[uint8]interface{})
	fmt.Printf("\x1b[1;33m")
	fmt.Printf("Cluster resources:")
	fmt.Printf("\x1b[0m\n")
	for _, selectedRsc := range resources {
		idAttr := selectedRsc.SelectAttr("id")
		if idAttr != nil {
			isISCSI := isISCSIEntry(idAttr.Value)
			isLU := false
			if isISCSI {
				isLU = isLogicalUnitEntry(idAttr.Value)
				fmt.Printf("\x1b[1;32m")
			}
			fmt.Printf("%-40s", idAttr.Value)
			if isISCSI {
				fmt.Printf(" \x1b[0;32m[iSCSI]")
				if isLU {
					fmt.Printf(" \x1b[1;34m[LU]")
				}
				if isISCSI {
					fmt.Printf("\x1b[0m")
				}
				fmt.Printf("\n")

				tidPrm := selectedRsc.FindElement("instance_attributes/nvpair[@name='tid']")
				if tidPrm != nil {
					tidAttr := tidPrm.SelectAttr("value")
					if tidAttr != nil {
						tid, err := strconv.ParseUint(tidAttr.Value, 10, 8)
						if err != nil {
							fmt.Printf("\x1b[1;31mWarning: Unparseable tid parameter '%s' for resource '%s'\x1b[0m\n", tidAttr.Value, idAttr.Value)
						}
						tidTable[uint8(tid)] = nil
					}
				}
			} else {
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("Warning: CIB primitive element has no attribute \x1b[1;32mname\x1b[0m\n")
		}
	}
	fmt.Printf("\n")

	if len(tidTable) > 0 {
		fmt.Printf("Allocated TIDs:\n")
		var tidList []uint8
		for tid, _ := range tidTable {
			tidList = append(tidList, tid)
		}
		sort.Sort(SortUint8(tidList))
		for _, tid := range tidList {
			fmt.Printf("    %d\n", tid)
		}
	} else {
		fmt.Printf("No TIDs allocated")
	}
	fmt.Printf("\n")

	return nil
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

func isISCSIEntry(id string) bool {
	return strings.Index(id, CRM_ISCSI_RSC_PREFIX) == 0
}

func isLogicalUnitEntry(id string) bool {
	result := false
	nameIdx := strings.LastIndexByte(id, '_')
	if nameIdx != -1 {
		name := id[nameIdx+1:]
		lunIdx := strings.Index(name, CRM_ISCSI_LU_NAME)
		if lunIdx == 0 {
			lunStr := name[lunIdx+len(CRM_ISCSI_LU_NAME):]
			_, err := strconv.ParseUint(lunStr, 10, 8)
			if err == nil {
				result = true
			}
		}
	}
	return result
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
