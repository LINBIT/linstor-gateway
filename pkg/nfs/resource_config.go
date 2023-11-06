package nfs

import (
	"errors"
	"fmt"
	apiconsts "github.com/LINBIT/golinstor"
	"github.com/icza/gog"
	log "github.com/sirupsen/logrus"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor/client"
	"github.com/google/uuid"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

const (
	ExportBasePath = "/srv/gateway-exports"
	DefaultNFSPort = 2049
	CurrentVersion = 1
)

var (
	UuidNFS      = uuid.NewSHA1(uuid.Nil, []byte("nfs.gateway.linstor.linbit.com"))
	AllowAllCidr = []common.IpCidr{
		{IPNet: net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}},
		{IPNet: net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)}},
	}
)

// VolumeConfig adds an export path in addition to the LINSTOR common.VolumeConfig.
type VolumeConfig struct {
	common.VolumeConfig
	ExportPath string `json:"export_path"`
}

// rootedPath returns a cleaned up path, rooted at /.
func rootedPath(path string) string {
	return filepath.Clean(filepath.Join("/", path))
}

// ExportPath returns the full path under which the resource is exported.
func ExportPath(rsc *ResourceConfig, vol *VolumeConfig) string {
	return filepath.Join(ExportBasePath, rsc.Name, vol.ExportPath)
}

type ResourceConfig struct {
	Name          string                `json:"name"`
	ServiceIP     common.IpCidr         `json:"service_ip,omitempty"`
	AllowedIPs    []common.IpCidr       `json:"allowed_ips,omitempty"`
	ResourceGroup string                `json:"resource_group"`
	Volumes       []VolumeConfig        `json:"volumes"`
	Status        common.ResourceStatus `json:"status"`
	GrossSize     bool                  `json:"gross_size"`
}

const (
	fsAgentName     = "fs_%d"
	exportAgentName = "export_%d_%d"
)

func FromPromoter(cfg *reactor.PromoterConfig, definition *client.ResourceDefinition, volumeDefinition []client.VolumeDefinition) (*ResourceConfig, error) {
	r := &ResourceConfig{}

	var rscCfg *reactor.PromoterResourceConfig
	r.Name, rscCfg = cfg.FirstResource()
	if rscCfg == nil {
		return nil, fmt.Errorf("promoter config without resource")
	}

	if definition != nil {
		r.ResourceGroup = definition.ResourceGroupName
	}

	if len(rscCfg.Start) < 1 {
		return nil, errors.New("expected at least one resource agent to be configured")
	}

	var numPortblocks, numPortunblocks int
	for _, entry := range rscCfg.Start {
		switch agent := entry.(type) {
		case *reactor.ResourceAgent:
			switch agent.Type {
			case "ocf:heartbeat:portblock":
				switch agent.Attributes["action"] {
				case "block":
					numPortblocks++
				case "unblock":
					numPortunblocks++
				}
			case "ocf:heartbeat:Filesystem":
				vol, err := parseVolume(agent, volumeDefinition, r.Name)
				if err != nil {
					log.Warnf("ignoring invalid resource agent: %v", err)
					continue
				}

				r.Volumes = append(r.Volumes, *vol)
			case "ocf:heartbeat:exportfs":
				cidr, err := cidrFromNfs(agent.Attributes["clientspec"])
				if err != nil {
					return nil, err
				}

				exists := false
				for i := range r.AllowedIPs {
					if r.AllowedIPs[i].String() == cidr.String() {
						exists = true
						break
					}
				}

				if !exists {
					r.AllowedIPs = append(r.AllowedIPs, cidr)
				}
			case "ocf:heartbeat:nfsserver":
				break
			case "ocf:heartbeat:IPaddr2":
				ip := net.ParseIP(agent.Attributes["ip"])
				if ip == nil {
					return nil, fmt.Errorf("malformed ip %s", agent.Attributes["ip"])
				}

				prefixLength, err := strconv.Atoi(agent.Attributes["cidr_netmask"])
				if err != nil {
					return nil, fmt.Errorf("failed to parse service ip prefix")
				}

				r.ServiceIP = common.ServiceIPFromParts(ip, prefixLength)
			default:
				return nil, errors.New(fmt.Sprintf("unexpected resource agent: %s", agent.Type))
			}
		case *reactor.SystemdService:
			// ignore systemd services for now
		}
	}

	if numPortblocks != numPortunblocks {
		return nil, fmt.Errorf("malformed configuration: got a different number of portblock and portunblock agents (%d vs %d)", numPortblocks, numPortunblocks)
	}

	if numPortblocks != 1 {
		return nil, fmt.Errorf("malformed configuration: got a different number of portblock agents (%d) than IPaddr2 agents (1)", numPortblocks)
	}

	return r, nil
}

// parseVolume converts a resource agent of the type "ocf:heartbeat:Filesystem"
// to a VolumeConfig.
// It associates the agent to a specific volume via its name (for details see
// findFilesystemAgentVolume). It also reconstructs the export path.
// Similarly to findFilesystemAgentVolume, it treats the agent name
// "fs_cluster_private" specially; this corresponds to the reserved cluster
// private volume with volume number 0.
func parseVolume(agent *reactor.ResourceAgent, volumes []client.VolumeDefinition, resName string) (*VolumeConfig, error) {
	vol, err := findFilesystemAgentVolume(volumes, agent)
	if err != nil {
		return nil, err
	}
	if vol == nil {
		return nil, fmt.Errorf("no volume definition found for resource agent %s", agent.Name)
	}

	dir := agent.Attributes["directory"]
	var exportPath string
	// the cluster private volume has no export path because it is not exported
	if agent.Name != common.ClusterPrivateVolumeAgentName {
		dirPrefix := filepath.Join(ExportBasePath, resName)
		if !strings.HasPrefix(dir, dirPrefix) {
			return nil, errors.New(fmt.Sprintf("export path %s not rooted in expected export path %s", dir, dirPrefix))
		}

		exportPath = rootedPath(dir[len(dirPrefix):])
	}

	var filesystem string
	if val, ok := vol.Props[apiconsts.NamespcFilesystem+"/Type"]; ok {
		filesystem = val
	}
	var rootOwner common.UidGid
	if val, ok := vol.Props[apiconsts.NamespcFilesystem+"/MkfsParams"]; ok {
		scanned, err := fmt.Sscanf(val, "-E root_owner=%d:%d", &rootOwner.Uid, &rootOwner.Gid)
		if scanned != 2 || err != nil {
			log.WithFields(log.Fields{
				"err":      err,
				"scanned":  scanned,
				"volume":   vol.VolumeNumber,
				"resource": resName,
			}).Warnf("invalid MkfsParams for volume: %q", val)
		}
	}
	if vol.VolumeNumber == nil {
		vol.VolumeNumber = gog.Ptr(int32(0))
	}
	return &VolumeConfig{
		VolumeConfig: common.VolumeConfig{
			Number:              int(*vol.VolumeNumber),
			SizeKiB:             vol.SizeKib,
			FileSystem:          filesystem,
			FileSystemRootOwner: rootOwner,
		},
		ExportPath: exportPath,
	}, nil
}

// findFilesystemAgentVolume searches "volumes" for the volume that is described
// by the ocf:heartbeat:Filesystem resource agent "agent". It does this by
// parsing the predefined agent name format (for example, fs_4 would correspond
// to volume ID 4).
// The agent name "fs_cluster_private" is handled specially, it always returns
// the reserved volume with the ID 0.
// If a parsing error occurs, an error is returned.
// If the volume is not found, a nil VolumeDefinition is returned.
func findFilesystemAgentVolume(volumes []client.VolumeDefinition, agent *reactor.ResourceAgent) (*client.VolumeDefinition, error) {
	var vol *client.VolumeDefinition

	if agent == nil || agent.Type != "ocf:heartbeat:Filesystem" {
		return nil, fmt.Errorf("invalid agent: %v", agent)
	}

	var volNr int
	if agent.Name == common.ClusterPrivateVolumeAgentName {
		volNr = 0
	} else {
		n, err := fmt.Sscanf(agent.Name, fsAgentName, &volNr)
		if n == 0 {
			return nil, fmt.Errorf("agent %s doesn't have expected name: %w", agent.Name, err)
		}
	}

	for i := range volumes {
		if volumes[i].VolumeNumber == nil {
			volumes[i].VolumeNumber = gog.Ptr(int32(0))
		}
		if int(*volumes[i].VolumeNumber) == volNr {
			vol = &volumes[i]
			break
		}
	}
	return vol, nil
}

func (r *ResourceConfig) VolumeConfig(number int) *common.Volume {
	var result *common.Volume

	for _, vol := range r.Volumes {
		if vol.Number == number {
			result = &common.Volume{
				Volume: vol.VolumeConfig,
			}
			break
		}
	}

	for _, volState := range r.Status.Volumes {
		if volState.Number == number {
			if result == nil {
				result = &common.Volume{}
			}

			result.Status = volState
			break
		}
	}

	return result
}

func (r *ResourceConfig) FillDefaults() {
	if r.ResourceGroup == "" {
		r.ResourceGroup = "DfltRscGrp"
	}

	for i := range r.Volumes {
		if r.Volumes[i].Number == 0 {
			// cluster private volume is supposed to have an empty export path
			continue
		}
		r.Volumes[i].ExportPath = rootedPath(r.Volumes[i].ExportPath)
	}

	if len(r.AllowedIPs) == 0 {
		r.AllowedIPs = AllowAllCidr
	}
}

func (r *ResourceConfig) Valid() error {
	if len(r.Name) < 2 {
		return common.ValidationError("nfs resource name to short (min. 2)")
	}

	if r.ServiceIP.IP() == nil {
		return common.ValidationError("missing service ip")
	}

	if r.ServiceIP.Mask == nil {
		return common.ValidationError("missing service ip prefix length")
	}

	sort.Slice(r.Volumes, func(i, j int) bool {
		return r.Volumes[i].Number < r.Volumes[j].Number
	})

	paths := make(map[string]struct{})

	for i := range r.Volumes {
		paths[r.Volumes[i].ExportPath] = struct{}{}

		if r.Volumes[i].Number < 0 {
			return common.ValidationError("volume numbers must start at 1")
		}

		if r.Volumes[i].SizeKiB <= 0 {
			return common.ValidationError("volume size must be positive")
		}

		if i > 0 && r.Volumes[i-1].Number == r.Volumes[i].Number {
			return common.ValidationError("volume numbers must be unique")
		}
	}

	if len(paths) != len(r.Volumes) {
		return common.ValidationError("nfs export paths must be unique")
	}

	return nil
}

func (r *ResourceConfig) Matches(o *ResourceConfig) bool {
	if r.Name != o.Name {
		return false
	}

	if r.ServiceIP.String() != o.ServiceIP.String() {
		return false
	}

	if r.ResourceGroup != o.ResourceGroup {
		return false
	}

	if len(r.AllowedIPs) != len(o.AllowedIPs) {
		return false
	}

	for i := range r.AllowedIPs {
		if r.AllowedIPs[i].String() != o.AllowedIPs[i].String() {
			return false
		}
	}

	if len(r.Volumes) != len(o.Volumes) {
		return false
	}

	for i := range r.Volumes {
		if r.Volumes[i].Number != o.Volumes[i].Number {
			return false
		}

		if r.Volumes[i].SizeKiB != o.Volumes[i].SizeKiB {
			return false
		}

		if r.Volumes[i].ExportPath != o.Volumes[i].ExportPath {
			return false
		}
	}

	return true
}

func (r *ResourceConfig) ID() string {
	return fmt.Sprintf(IDFormat, r.Name)
}

func (r *ResourceConfig) ToPromoter(deployment []client.ResourceWithVolumes) (*reactor.PromoterConfig, error) {
	if len(deployment) == 0 {
		return nil, errors.New("resource config is missing deployment information")
	}
	deployedRes := deployment[0]
	var agents []reactor.StartEntry
	resUuid := uuid.NewSHA1(UuidNFS, []byte(deployedRes.Uuid))

	log.Debugf("volumes: %+v", deployedRes.Volumes)

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:portblock",
		Name: "portblock",
		Attributes: map[string]string{
			"ip":       r.ServiceIP.IP().String(),
			"portno":   strconv.Itoa(DefaultNFSPort),
			"action":   "block",
			"protocol": "tcp",
		},
	})

	// volume 0 is reserved as the "cluster private" volume
	deployedClusterPrivateVol := deployedRes.Volumes[0]
	agents = append(agents, common.ClusterPrivateVolumeAgent(deployedClusterPrivateVol, r.Name))

	for i := 1; i < len(deployedRes.Volumes); i++ {
		vol := deployedRes.Volumes[i]
		resVol := r.Volumes[i]
		if int(vol.VolumeNumber) != resVol.Number {
			return nil, fmt.Errorf("inconsistent volumes, expected volume number %d, got %d", vol.VolumeNumber, resVol.Number)
		}

		dirPath := ExportPath(r, &resVol)

		agents = append(agents,
			&reactor.ResourceAgent{
				Type: "ocf:heartbeat:Filesystem",
				Name: fmt.Sprintf(fsAgentName, vol.VolumeNumber),
				Attributes: map[string]string{
					"device":    common.DevicePath(vol),
					"directory": dirPath,
					"fstype":    resVol.FileSystem,
					"run_fsck":  "no",
				},
			},
		)
	}

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:IPaddr2",
		Name: "service_ip",
		Attributes: map[string]string{
			"ip":           r.ServiceIP.IP().String(),
			"cidr_netmask": strconv.Itoa(r.ServiceIP.Prefix()),
		},
	})

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:nfsserver",
		Name: "nfsserver",
		Attributes: map[string]string{
			"nfs_ip":             r.ServiceIP.IP().String(),
			"nfs_shared_infodir": filepath.Join(common.ClusterPrivateVolumeMountPath, deployedRes.Name, "nfs"),
			"nfs_server_scope":   r.ServiceIP.IP().String(),
		},
	})

	for i := 1; i < len(deployedRes.Volumes); i++ {
		vol := deployedRes.Volumes[i]
		resVol := r.Volumes[i]
		if int(vol.VolumeNumber) != resVol.Number {
			return nil, fmt.Errorf("inconsistent volumes, expected volume number %d, got %d", vol.VolumeNumber, resVol.Number)
		}

		fsid := uuid.NewSHA1(resUuid, []byte(vol.Uuid))

		dirPath := ExportPath(r, &resVol)

		for j := range r.AllowedIPs {
			agents = append(agents, &reactor.ResourceAgent{
				Type: "ocf:heartbeat:exportfs",
				Name: fmt.Sprintf(exportAgentName, vol.VolumeNumber, j),
				Attributes: map[string]string{
					"directory":  dirPath,
					"fsid":       fsid.String(),
					"clientspec": nfsFormatCidr(&r.AllowedIPs[j]),
					"options":    "rw,all_squash,anonuid=0,anongid=0",
				},
			})
		}
	}

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:portblock",
		Name: "portunblock",
		Attributes: map[string]string{
			"ip":         r.ServiceIP.IP().String(),
			"portno":     strconv.Itoa(DefaultNFSPort),
			"action":     "unblock",
			"protocol":   "tcp",
			"tickle_dir": filepath.Join(common.ClusterPrivateVolumeMountPath, deployedRes.Name),
		},
	})

	return &reactor.PromoterConfig{
		Resources: map[string]reactor.PromoterResourceConfig{
			r.Name: {
				Runner:              "systemd",
				Start:               agents,
				StopServicesOnExit:  true,
				OnDrbdDemoteFailure: "reboot-immediate",
				TargetAs:            "BindsTo",
			},
		},
		Metadata: reactor.PromoterMetadata{
			LinstorGatewaySchemaVersion: CurrentVersion,
		},
	}, nil
}

func cidrFromNfs(nfs string) (common.IpCidr, error) {
	parts := strings.SplitN(nfs, "/", 2)
	if len(parts) != 2 {
		return common.IpCidr{}, errors.New(fmt.Sprintf("expected at least one / in nfs export control %s", nfs))
	}

	ip := net.ParseIP(strings.Trim(parts[0], "[]"))
	if ip == nil {
		return common.IpCidr{}, errors.New(fmt.Sprintf("failed to parse nfs export control %s as ip", parts[0]))
	}

	if strings.Contains(parts[1], ".") {
		mask := net.ParseIP(parts[1])
		if mask == nil {
			return common.IpCidr{}, errors.New(fmt.Sprintf("failed to parse nfs export control %s as mask", parts[1]))
		}

		return common.IpCidr{
			IPNet: net.IPNet{
				IP:   ip,
				Mask: net.IPMask(mask),
			},
		}, nil
	} else {
		prefix, err := strconv.Atoi(parts[1])
		if err != nil {
			return common.IpCidr{}, errors.New(fmt.Sprintf("failed to parse cidr prefix length: %s", parts[1]))
		}

		return common.ServiceIPFromParts(ip, prefix), nil
	}
}
func nfsFormatCidr(n *common.IpCidr) string {
	if n.Prefix() == 0 && n.IP().To4() != nil {
		// For some reason, nfs does not understand a.b.c.d/0. Instead, we have to use a.b.c.d/0.0.0.0.
		mask := n.IPNet.Mask
		return fmt.Sprintf("%s/%d.%d.%d.%d", n.IP().String(), mask[0], mask[1], mask[2], mask[3])
	}

	if n.IP().To4() == nil {
		return fmt.Sprintf("[%s]/%d", n.IP().String(), n.Prefix())
	}

	return n.String()
}
