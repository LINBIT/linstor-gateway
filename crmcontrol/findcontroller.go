package crmcontrol

import (
	"errors"
	"fmt"
	"net"
)

func findLinstorControllerName() (string, error) {
	doc, err := ReadConfiguration()
	if err != nil {
		return "", err
	}

	xpath := fmt.Sprintf("%s/%s/%s", CIB_NODE_STATUS_XPATH, CIB_TAG_LRM, CIB_TAG_LRM_RSCLST)
	for _, lrm_resources := range doc.FindElements(xpath) {
		for _, lrm_resource := range lrm_resources.SelectElements(CIB_TAG_LRM_RSC) {
			typ := lrm_resource.SelectAttrValue("type", "")
			class := lrm_resource.SelectAttrValue("class", "")
			if typ == CRM_TYPE_LINSTOR_CTRL && class == "systemd" {
				if lrm_rsc_op := lrm_resource.SelectElement(CIB_TAG_LRM_RSC_OP); lrm_rsc_op != nil {
					on_node := lrm_rsc_op.SelectAttrValue("on_node", "")
					operation := lrm_rsc_op.SelectAttrValue("operation", "")
					if on_node != "" && operation == "start" {
						return on_node, nil
					}
				}
			}
		}
	}

	return "", errors.New("Could not find the 'linstor-controller' in the CIB")
}

func FindLinstorController() (net.IP, error) {
	hostname, err := findLinstorControllerName()
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	return ips[0], nil
}
