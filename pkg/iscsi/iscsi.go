// Package iscsi combines LINSTOR operations and the CRM operations to create highly available iSCSI targets.
package iscsi

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
	"strings"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"

	xmltree "github.com/beevik/etree"
)

// Default port for an iSCSI portal
const DFLT_ISCSI_PORTAL_PORT = 3260

// ISCSI combines the information needed to create highly-available iSCSI targets.
// It contains a iSCSI target configuration and a LINSTOR configuration.
type ISCSI struct {
	Target  Target                 `json:"target,omitempty"`
	Linstor linstorcontrol.Linstor `json:"linstor,omitempty"`
}

type LUN struct {
	ID      uint8  `json:"id,omitempty"`
	SizeKiB uint64 `json:"size_kib,omitempty"`
}

const (
	// This format is currently dictated by the iSCSI target backend,
	// specifically the rtslib-fb library.
	// A notable difference in this implementation (which also differs from
	// RFC3720, where the IQN format is defined) is that we require the
	// "unique" part after the colon to be present.
	//
	// See also the source code of rtslib-fb for the original regex:
	// https://github.com/open-iscsi/rtslib-fb/blob/b5be390be961/rtslib/utils.py#L384
	regexIQN = `iqn\.\d{4}-[0-1][0-9]\..*\..*`

	// This format is mandated by LINSTOR. Since we use the unique part
	// directly for LINSTOR resource names, it needs to be compliant.
	regexResourceName = `[[:alpha:]][[:alnum:]]+`

	regexWWN = `^` + regexIQN + `:` + regexResourceName + `$`
)

// TargetConfig contains the information necessary for iSCSI targets.
type TargetConfig struct {
	Name      string `json:"name,omitempty"`
	IQN       string `json:"iqn,omitempty"`
	LUNs      []*LUN `json:"luns,omitempty"`
	ServiceIP net.IP `json:"service_ip,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Portals   string `json:"portals,omitempty"`
}

// Target wraps a TargetConfig
type Target struct {
	TargetConfig
}

func CheckIQN(iqn string) error {
	if strings.ContainsAny(iqn, "_ ") {
		return errors.New("IQN cannot contain the characters '_' (underscore) or ' ' (space)")
	}

	matched, err := regexp.MatchString(regexWWN, iqn)
	if err != nil {
		return err
	} else if !matched {
		return errors.New("Given IQN does not match the regular expression: " + regexWWN)
	}

	return nil
}

func NewTarget(cfg TargetConfig) (Target, error) {
	if err := CheckIQN(cfg.IQN); err != nil {
		return Target{}, err
	}

	return Target{cfg}, nil
}

func NewTargetMust(cfg TargetConfig) Target {
	t, err := NewTarget(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// CreateResource creates a new highly available iSCSI target
func (i *ISCSI) CreateResource() error {
	targetName, err := ExtractTargetName(i.Target.IQN)
	if err != nil {
		return err
	}

	for _, lu := range i.Target.LUNs {
		// Read the current configuration from the CRM
		docRoot, err := crmcontrol.ReadConfiguration()
		if err != nil {
			return err
		}
		// Find resources, allocated target IDs, etc.
		config, err := crmcontrol.ParseConfiguration(docRoot)
		if err != nil {
			return err
		}

		// Find a free target ID number using the set of allocated target IDs
		freeTid, ok := config.TIDs.GetFree(1, math.MaxInt16)
		if !ok {
			return errors.New("Failed to allocate a target ID for the new iSCSI target")
		}

		// Create a LINSTOR resource definition, volume definition and associated resources
		i.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, lu.ID)
		i.Linstor.SizeKiB = lu.SizeKiB
		res, err := i.Linstor.CreateVolume()
		if err != nil {
			return fmt.Errorf("LINSTOR volume operation failed, error: %v", err)
		}

		// Create CRM resources and constraints for the iSCSI services
		err = crmcontrol.CreateCrmLu(
			res.StorageNodeList,
			targetName,
			i.Target.ServiceIP,
			i.Target.IQN,
			lu.ID,
			res.DevicePath,
			i.Target.Username,
			i.Target.Password,
			i.Target.Portals,
			int16(freeTid),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteResource deletes a highly available iSCSI target
func (i *ISCSI) DeleteResource() error {
	targetName, err := ExtractTargetName(i.Target.IQN)
	if err != nil {
		return err
	}

	for _, lu := range i.Target.LUNs {
		// Delete the CRM resources for iSCSI LU, target, service IP addres, etc.
		err = crmcontrol.DeleteCrmLu(targetName, lu.ID)
		if err != nil {
			return err
		}

		// Delete the LINSTOR resource definition
		i.Linstor.ResourceName = linstorcontrol.ResourceNameFromLUN(targetName, lu.ID)
		err = i.Linstor.DeleteVolume()
		if err != nil {
			return err
		}
	}
	return nil
}

// StartResource starts an existing iSCSI resource.
func (i *ISCSI) StartResource() error {
	return i.modifyResourceTargetRole(true)
}

// StopResource stops an existing iSCSI resource.
func (i *ISCSI) StopResource() error {
	return i.modifyResourceTargetRole(false)
}

// ProbeResource gets information about an existing iSCSI resource.
// It returns a resource state map and an error.
func (i *ISCSI) ProbeResource() (crmcontrol.ResourceRunState, error) {
	targetName, err := ExtractTargetName(i.Target.IQN)
	if err != nil {
		return crmcontrol.ResourceRunState{}, err
	}

	luns := make([]uint8, len(i.Target.LUNs))
	for i, lu := range i.Target.LUNs {
		luns[i] = lu.ID
	}

	return crmcontrol.ProbeResource(targetName, luns)
}

// ListResources lists existing iSCSI targets.
//
// Returns: CIB XML document tree, slice of Targets, error object
func ListResources() (*xmltree.Document, []*Target, error) {
	docRoot, err := crmcontrol.ReadConfiguration()
	if err != nil {
		return nil, nil, err
	}

	config, err := crmcontrol.ParseConfiguration(docRoot)
	if err != nil {
		return nil, nil, err
	}

	targets := make([]*Target, 0)

	// first, "convert" all targets
	for _, t := range config.Targets {
		targetCfg := TargetConfig{
			Name:     t.ID,
			IQN:      t.IQN,
			LUNs:     make([]*LUN, 0),
			Username: t.Username,
			Password: t.Password,
			Portals:  t.Portals,
		}

		target, err := NewTarget(targetCfg)
		if err != nil {
			return nil, nil, err
		}

		targets = append(targets, &target)
	}

	// then, "convert" and link LUs
	for _, l := range config.LUs {
		lun := &LUN{
			ID: l.LUN,
		}

		// link to the correct target
		for _, t := range targets {
			if t.IQN == l.Target.IQN {
				t.LUNs = append(t.LUNs, lun)
				break
			}
		}
	}

	return docRoot, targets, nil
}

// modifyResourceTargetRole modifies the role of an existing iSCSI resource.
func (i *ISCSI) modifyResourceTargetRole(startFlag bool) error {
	targetName, err := ExtractTargetName(i.Target.IQN)
	if err != nil {
		return errors.New("Invalid IQN format: Missing ':' separator and target name")
	}

	luns := make([]uint8, len(i.Target.LUNs))
	for i, lu := range i.Target.LUNs {
		luns[i] = lu.ID
	}
	if startFlag {
		crmcontrol.StartCrmResource(targetName, luns)
	} else {
		crmcontrol.StopCrmResource(targetName, luns)
	}

	return nil
}

// ExtractTargetName extracts the target name from an IQN string.
// e.g., in "iqn.2019-07.org.demo.filserver:filestorage", the "filestorage" part.
func ExtractTargetName(iqn string) (string, error) {
	spl := strings.Split(iqn, ":")
	if len(spl) != 2 {
		return "", errors.New("Malformed argument '" + iqn + "'")
	}
	return spl[1], nil
}
