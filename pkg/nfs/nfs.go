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

// Top-level object comprising an NFS export configuration (NfsConfig) with
// its associated Linstor resource/volume configuration (Linstor)
type NfsResource struct {
	Nfs     nfsbase.NfsConfig      `json:"nfsexport,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

type NfsListItem struct {
	ResourceName	string
	LinstorRsc      linstorcontrol.Linstor
	Mountpoint      crmcontrol.FSMount
	NfsExport       crmcontrol.ExportFS
	ServiceIP       crmcontrol.IP
}

func (nfsRsc *NfsResource) CreateResource() error {
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
	nfsRsc.Linstor.ResourceName = nfsRsc.Nfs.ResourceName
	nfsRsc.Linstor.SizeKiB = nfsRsc.Nfs.SizeKiB
	rsc, err := nfsRsc.Linstor.CreateVolume()
	if err != nil {
		return fmt.Errorf("LINSTOR volume operations failed, error: %v", err)
	}

	// Create the NFS export directory
	log.Debug("nfs.go CreateResource: Creating export directory")
	directory := nfsbase.NfsBasePath + "/" + nfsRsc.Nfs.ResourceName
	err = os.Mkdir(directory, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	log.Debug("nfs.go CreateResource: Creating CRM resources and constraints")
	// Create CRM resources and constraints for the NFS export
	err = crmcontrol.CreateNfs(nfsRsc.Nfs, rsc.StorageNodeList, rsc.DevicePath, directory)
	if err != nil {
		return err
	}

	log.Debug("nfs.go CreateResource: Successful")
	return nil
}

func (nfsRsc *NfsResource) DeleteResource() error {
	var errors []string

	// Delete the CRM resources
	if err := crmcontrol.DeleteNfs(nfsRsc.Nfs); err != nil {
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

func ListResources() ([]NfsListItem, error) {
	items := make([]NfsListItem, 0)

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

	for _, nfsExport := range config.NfsExports {
		rscName, isExport := getRscNameFromNfsExport(nfsExport)
		if isExport {
			mountpoint, haveMountpoint := mountpointMap["p_nfs_" + rscName + "_fs"]
			svcIP, haveSvcIP := svcIPMap["p_nfs_" + rscName + "_ip"]

			if haveMountpoint && haveSvcIP {
				entry := NfsListItem{
					ResourceName: rscName,
					Mountpoint:   *mountpoint,
					NfsExport:    *nfsExport,
					ServiceIP:    *svcIP,
				}
				items = append(items, entry)
			}
		}
	}

	return items, nil
}

func (nfsRsc *NfsResource) StartResource() error {
	return nfsRsc.modifyResourceTargetRole(true)
}

func (nfsRsc *NfsResource) StopResource() error {
	return nfsRsc.modifyResourceTargetRole(false)
}

func (nfsRsc *NfsResource) ProbeResource() (crmcontrol.NfsRunState, error) {
	return crmcontrol.ProbeNfsResource(nfsRsc.Nfs.ResourceName)
}

func (nfsRsc *NfsResource) modifyResourceTargetRole(flag bool) error {
	// TODO: Implement NFS CRM resource start/stop
	return nil
}

func getRscNameFromNfsExport(nfsExport *crmcontrol.ExportFS) (string, bool) {
	var rscName string
	var isExport bool = false
	id := nfsExport.ID
	if strings.HasPrefix(id, "p_nfs_") && strings.HasSuffix(id, "_exp") {
		rscName = id[6:len(id) - 4]
		isExport = true
	}
	return rscName, isExport
}
