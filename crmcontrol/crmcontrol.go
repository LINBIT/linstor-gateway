package crmcontrol

import "fmt"
import "strings"
import "strconv"
import "github.com/LINBIT/linstor-remote-storage/templateproc"
import "github.com/LINBIT/linstor-remote-storage/extcmd"
import "github.com/LINBIT/linstor-remote-storage/debug"

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
