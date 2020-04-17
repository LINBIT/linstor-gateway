// Package nfs combines LINSTOR operations and the CRM operations to create highly available NFS exports.
package nfs

import (
	"fmt"
	"os"
	"strings"
)

import (
	log "github.com/sirupsen/logrus"
	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/nfsbase"
)

// Top-level object comprising an NFS export configuration (NFSConfig) with
// its associated Linstor resource/volume configuration (Linstor)
type NFSResource struct {
	NFS     nfsbase.NFSConfig      `json:"nfsexport,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

type NFSListItem struct {
	ResourceName	string
	LinstorRsc      linstorcontrol.Linstor
	Mountpoint      crmcontrol.FSMount
	NFSExport       crmcontrol.ExportFS
	ServiceIP       crmcontrol.IP
}

func (nfsRsc *NFSResource) CreateResource() error {
	log.Debug("nfs.go CreateResource: Reading CIB")
	var cibObj cib.CIB
	// Read the current configuration from the CRM
	err := cibObj.ReadConfiguration()
	if err != nil {
		return err
	}

	// TODO: Do something with the CRM configuration
	// crmConfig, err := crmcontrol.ParseConfiguration(cibObj.Doc)
	// if err != nil {
	// 	return err
	// }

	log.Debug("nfs.go CreateResource: Creating LINSTOR resource")
	// Create a LINSTOR resource definition, volume definition and associated resources
	nfsRsc.Linstor.ResourceName = nfsRsc.NFS.ResourceName
	nfsRsc.Linstor.SizeKiB = nfsRsc.NFS.SizeKiB
	rsc, err := nfsRsc.Linstor.CreateVolume()
	if err != nil {
		return fmt.Errorf("LINSTOR volume operations failed, error: %v", err)
	}

	// Create the NFS export directory
	log.Debug("nfs.go CreateResource: Creating export directory")
	directory := nfsbase.NFSBasePath + "/" + nfsRsc.NFS.ResourceName
	err = os.Mkdir(directory, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	log.Debug("nfs.go CreateResource: Creating CRM resources and constraints")
	// Create CRM resources and constraints for the NFS export
	err = crmcontrol.CreateNFS(nfsRsc.NFS, rsc.StorageNodeList, rsc.DevicePath, directory)
	if err != nil {
		return err
	}

	log.Debug("nfs.go CreateResource: Successful")
	return nil
}

func (nfsRsc *NFSResource) DeleteResource() error {
	var errors []string

	// Delete the CRM resources
	if err := crmcontrol.DeleteNFS(nfsRsc.NFS); err != nil {
		errors = append(errors, err.Error())
	}

	// Delete the LINSTOR resource definition
	if err := nfsRsc.Linstor.DeleteVolume(); err != nil {
		errors = append(errors, err.Error())
	}

	// TODO: Check the validity of the mount point directory path
	//       Linstor may have checked it already, it may be
	//       sufficient to rely on LINSTOR's check
	// TODO: Delete mount point directory

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	return nil
}

func ListResources() ([]NFSListItem, error) {
	items := make([]NFSListItem, 0)

	var cibObj cib.CIB
	err := cibObj.ReadConfiguration()
	if err != nil {
		return nil, err
	}

	config, err := crmcontrol.ParseConfiguration(cibObj.Doc)
	if err != nil {
		return nil, err
	}

	mountpointMap := make(map[string]*crmcontrol.FSMount)
	for _, item := range config.Mountpoints {
		mountpointMap[item.ID] = item
	}
	svcIPMap := make(map[string]*crmcontrol.IP)
	for _, item := range config.IPs {
		svcIPMap[item.ID] = item
	}

	for _, nfsExport := range config.NFSExports {
		rscName, isExport := getRscNameFromNFSExport(nfsExport)
		if isExport {
			mountpoint, haveMountpoint := mountpointMap["p_nfs_" + rscName + "_fs"]
			svcIP, haveSvcIP := svcIPMap["p_nfs_" + rscName + "_ip"]

			if haveMountpoint && haveSvcIP {
				entry := NFSListItem{
					ResourceName: rscName,
					Mountpoint:   *mountpoint,
					NFSExport:    *nfsExport,
					ServiceIP:    *svcIP,
				}
				items = append(items, entry)
			}
		}
	}

	return items, nil
}

func (nfsRsc *NFSResource) StartResource() error {
	return nfsRsc.modifyResourceTargetRole(true)
}

func (nfsRsc *NFSResource) StopResource() error {
	return nfsRsc.modifyResourceTargetRole(false)
}

func (nfsRsc *NFSResource) ProbeResource() (crmcontrol.NFSRunState, error) {
	return crmcontrol.ProbeNFSResource(nfsRsc.NFS.ResourceName)
}

func (nfsRsc *NFSResource) modifyResourceTargetRole(flag bool) error {
	// TODO: Implement NFS CRM resource start/stop
	return nil
}

func getRscNameFromNFSExport(nfsExport *crmcontrol.ExportFS) (string, bool) {
	var rscName string
	var isExport bool = false
	id := nfsExport.ID
	if strings.HasPrefix(id, "p_nfs_") && strings.HasSuffix(id, "_exp") {
		rscName = id[6:len(id) - 4]
		isExport = true
	}
	return rscName, isExport
}
