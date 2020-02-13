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

// isRunning returns whether or not a resource is running or not based on the
// contents of an lrm_resource_op. If it is either a start action or a successful
// monitor action, the resource is considered running.
func isRunning(on_node, operation, rcCode string) bool {
	return on_node != "" && (operation == "start" || (operation == "monitor" && rcCode == "0"))
}

func findLinstorControllerName(doc *xmltree.Document) string {
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
					rcCode := lrm_rsc_op.SelectAttrValue("rc-code", "")
					log.Tracef("Looking at rsc_op with operation %s on node %s", operation, on_node)
					if isRunning(on_node, operation, rcCode) {
						log.Tracef("Found LINSTOR controller on node %s", on_node)
						return on_node
					}
				}
			}
		}
	}

	return ""
}

// FindLinstorController searches the CIB configuration for a LINSTOR controller IP.
func FindLinstorController() (net.IP, error) {
	var c cib.CIB
	err := c.ReadConfiguration()
	if err != nil {
		return nil, err
	}

	hostname := findLinstorControllerName(c.Doc)
	if hostname == "" {
		return nil, errors.New("Could not find the 'linstor-controller' in the CIB")
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	return ips[0], nil
}
