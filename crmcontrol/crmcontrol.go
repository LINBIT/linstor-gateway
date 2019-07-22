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

type CrmConfiguration struct {
	TargetList []string
	LuList     []string
	TidSet     TargetIdSet
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

	if err != nil {
		return err
	}

	return nil
}

func DeleteCrmLu(
	iscsiTargetName string,
	lun uint8,
) error {
	luName := CRM_ISCSI_LU_NAME + strconv.Itoa(int(lun))

	crmLu := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_" + luName
	crmTgt := CRM_ISCSI_RSC_PREFIX + iscsiTargetName
	crmSvcIp := CRM_ISCSI_RSC_PREFIX + iscsiTargetName + "_ip"

	cmd, _, err := extcmd.PipeToExtCmd(
		"cibadmin",
		[]string{
			"--delete",
			"--xpath=/cib/configuration/resources/primitive[@id='" + crmLu + "']",
		},
	)
	stdoutLines, stderrLines, err := cmd.WaitForExtCmd()
	printCmdOutput(stdoutLines, stderrLines)
	if err != nil {
		return err
	}

	cmd, _, err = extcmd.PipeToExtCmd(
		"cibadmin",
		[]string{
			"--delete",
			"--xpath=/cib/configuration/resources/primitive[@id='" + crmTgt + "']",
		},
	)
	stdoutLines, stderrLines, err = cmd.WaitForExtCmd()
	printCmdOutput(stdoutLines, stderrLines)
	if err != nil {
		return err
	}

	cmd, _, err = extcmd.PipeToExtCmd(
		"cibadmin",
		[]string{
			"--delete",
			"--xpath=/cib/configuration/resources/primitive[@id='" + crmSvcIp + "']",
		},
	)
	stdoutLines, stderrLines, err = cmd.WaitForExtCmd()
	printCmdOutput(stdoutLines, stderrLines)
	if err != nil {
		return err
	}

	return nil
}

func ReadConfiguration() (CrmConfiguration, error) {
	config := CrmConfiguration{TidSet: NewTargetIdSet()}
	docRoot := xmltree.NewDocument()

	cmd, _, err := extcmd.PipeToExtCmd("cibadmin", []string{"--query"})
	if err != nil {
		return config, err
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
		return config, err
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

	fmt.Printf("\x1b[1;33m")
	fmt.Printf("Cluster resources:")
	fmt.Printf("\x1b[0m\n")
	for _, selectedRsc := range resources {
		idAttr := selectedRsc.SelectAttr("id")
		if idAttr != nil {
			crmRscName := idAttr.Value
			isISCSI := strings.Index(idAttr.Value, CRM_ISCSI_RSC_PREFIX) == 0
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
			}
			if isLu {
				config.LuList = append(config.LuList, crmRscName)
			}
			fmt.Printf("%-40s", idAttr.Value)
			if isISCSI {
				fmt.Printf(" \x1b[0;32m[iSCSI]")
				if isTarget {
					fmt.Printf(" \x1b[1;33m[TARGET]")
				} else if isLu {
					fmt.Printf(" \x1b[1;34m[LU]")
				}
				fmt.Printf("\x1b[0m")
			} else {
				fmt.Printf(" \x1b[0;35m[other]\x1b[0m")
			}
			fmt.Printf("\n")

			tidEntry := selectedRsc.FindElement("instance_attributes/nvpair[@name='tid']")
			if tidEntry != nil {
				tidAttr := tidEntry.SelectAttr("value")
				if tidAttr != nil {
					tid, err := strconv.ParseUint(tidAttr.Value, 10, 8)
					if err != nil {
						fmt.Printf("\x1b[1;31mWarning: Unparseable tid parameter '%s' for resource '%s'\x1b[0m\n", tidAttr.Value, idAttr.Value)
					}
					config.TidSet.Insert(uint8(tid))
				}
			}
		} else {
			fmt.Printf("Warning: CIB primitive element has no attribute \x1b[1;32mname\x1b[0m\n")
		}
	}
	fmt.Printf("\n")

	if config.TidSet.GetSize() > 0 {
		fmt.Printf("Allocated TIDs:\n")
		tidIter := config.TidSet.Iterator()
		for tid, isValid := tidIter.Next(); isValid; tid, isValid = tidIter.Next() {
			fmt.Printf("    %d\n", tid)
		}
	} else {
		fmt.Printf("No TIDs allocated")
	}
	fmt.Printf("\n")

	freeTid, haveFreeTid := GetFreeTargetId(config.TidSet.ToSortedArray())
	if haveFreeTid {
		fmt.Printf("Next free TID: %d\n", int(freeTid))
	} else {
		fmt.Printf("No free TIDs")
	}
	fmt.Printf("\n")

	return config, nil
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
