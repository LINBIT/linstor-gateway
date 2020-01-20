package crmcontrol

import (
	"errors"
	"fmt"
	"net"

	"github.com/LINBIT/gopacemaker/cib"
	xmltree "github.com/beevik/etree"
	log "github.com/sirupsen/logrus"
)

const (
	cibTagLrm        = "lrm"
	cibTagLrmRsclist = "lrm_resources"
	cibTagLrmRsc     = "lrm_resource"
	cibTagLrmRscOp   = "lrm_rsc_op"
)

const (
	crmTypeLinstorCtrl = "linstor-controller"
)

func findLinstorControllerName(doc *xmltree.Document) (string, error) {
	xpath := fmt.Sprintf("%s/%s/%s", cibNodeStatusXpath, cibTagLrm, cibTagLrmRsclist)
	log.Tracef("Looking for linstor controller in %s", xpath)
	for _, lrm_resources := range doc.FindElements(xpath) {
		log.Tracef("Looking at element %v", lrm_resources)
		for _, lrm_resource := range lrm_resources.SelectElements(cibTagLrmRsc) {
			typ := lrm_resource.SelectAttrValue("type", "")
			class := lrm_resource.SelectAttrValue("class", "")
			log.Tracef("Looking at resource with type %s and class %s", typ, class)
			if typ == crmTypeLinstorCtrl && class == "systemd" {
				if lrm_rsc_op := lrm_resource.SelectElement(cibTagLrmRscOp); lrm_rsc_op != nil {
					on_node := lrm_rsc_op.SelectAttrValue("on_node", "")
					operation := lrm_rsc_op.SelectAttrValue("operation", "")
					log.Tracef("Looking at rsc_op with operation %s on node %s", operation, on_node)
					if on_node != "" && operation == "start" {
						log.Tracef("Found LINSTOR controller on node %s", on_node)
						return on_node, nil
					}
				}
			}
		}
	}

	return "", errors.New("Could not find the 'linstor-controller' in the CIB")
}

// FindLinstorController searches the CIB configuration for a LINSTOR controller IP.
func FindLinstorController() (net.IP, error) {
	var c cib.CIB
	err := c.ReadConfiguration()
	if err != nil {
		return nil, err
	}

	hostname, err := findLinstorControllerName(c.Doc)
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	return ips[0], nil
}
