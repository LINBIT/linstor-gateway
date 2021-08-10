package nfs

import (
	"errors"
	"fmt"
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
}

const (
	fsAgentName     = "fs_%d"
	exportAgentName = "export_%d_%d"
)

func FromPromoter(cfg *reactor.PromoterConfig, definition *client.ResourceDefinition, volumeDefinition []client.VolumeDefinition) (*ResourceConfig, error) {
	r := &ResourceConfig{}
	var res string
	n, err := fmt.Sscanf(cfg.ID, IDFormat, &res)
	if n != 1 {
		return nil, fmt.Errorf("failed to parse id into resource name: %w", err)
	}

	r.Name = res
	r.ResourceGroup = definition.ResourceGroupName

	if len(cfg.Resources) != 1 {
		return nil, errors.New(fmt.Sprintf("promoter config without exactly 1 resource (has %d)", len(cfg.Resources)))
	}

	var rscCfg reactor.PromoterResourceConfig
	for _, v := range cfg.Resources {
		rscCfg = v
	}

	if len(rscCfg.Start) < 1 {
		return nil, errors.New("expected at least one resource agent to be configured")
	}

	for i := range rscCfg.Start[:len(rscCfg.Start)-1] {
		agent := &rscCfg.Start[i]

		switch agent.Type {
		case "ocf:heartbeat:Filesystem":
			dir := agent.Attributes["directory"]

			dirPrefix := filepath.Join(ExportBasePath, r.Name)
			if !strings.HasPrefix(dir, dirPrefix) {
				return nil, errors.New(fmt.Sprintf("export path not rooted in expected export path %s", dir))
			}

			var volNr int
			n, err = fmt.Sscanf(agent.Name, fsAgentName, &volNr)
			if n == 0 {
				return nil, fmt.Errorf("agent %s doesn't have expected name: %w", agent.Name, err)
			}

			var vol *client.VolumeDefinition
			for i := range volumeDefinition {
				if int(volumeDefinition[i].VolumeNumber) == volNr {
					vol = &volumeDefinition[i]
					break
				}
			}

			if vol == nil {
				return nil, fmt.Errorf("no volume definition for volume number %d", volNr)
			}

			r.Volumes = append(r.Volumes, VolumeConfig{
				VolumeConfig: common.VolumeConfig{
					Number:  volNr,
					SizeKiB: vol.SizeKib,
				},
				ExportPath: rootedPath(dir[len(dirPrefix):]),
			})
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
		default:
			return nil, errors.New(fmt.Sprintf("unexpected resource agent: %s", agent.Type))
		}
	}

	ipAgent := &rscCfg.Start[len(rscCfg.Start)-1]
	if ipAgent.Type != "ocf:heartbeat:IPaddr2" {
		return nil, errors.New(fmt.Sprintf("expected 'ocf:heartbeat:IPaddr2' agent, got '%s' instead", ipAgent.Type))
	}

	ip := net.ParseIP(ipAgent.Attributes["ip"])
	if ip == nil {
		return nil, fmt.Errorf("malformed ip %s", ipAgent.Attributes["ip"])
	}

	prefixLength, err := strconv.Atoi(ipAgent.Attributes["cidr_netmask"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse service ip prefix")
	}

	r.ServiceIP = common.ServiceIPFromParts(ip, prefixLength)

	return r, nil
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
			r.Volumes[i].Number = i + 1
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

	var agents []reactor.ResourceAgent

	resUuid := uuid.NewSHA1(UuidNFS, []byte(deployment[0].Uuid))

	for i, vol := range deployment[0].Volumes {
		if int(vol.VolumeNumber) != r.Volumes[i].Number {
			return nil, fmt.Errorf("inconsistent volumes, expected volume number %d, got %d", vol.VolumeNumber, r.Volumes[i].Number)
		}

		devPath := vol.DevicePath
		for k, v := range vol.Props {
			if strings.HasPrefix(k, "Satellite/Device/Symlinks/") {
				devPath = v
			}

			// Prefer the "by-res" symlinks
			if strings.Contains(v, "/by-res/") {
				break
			}
		}
		dirPath := ExportPath(r, &r.Volumes[i])

		fsid := uuid.NewSHA1(resUuid, []byte(vol.Uuid))

		agents = append(agents,
			reactor.ResourceAgent{
				Type: "ocf:heartbeat:Filesystem",
				Name: fmt.Sprintf(fsAgentName, vol.VolumeNumber),
				Attributes: map[string]string{
					"device":    devPath,
					"directory": dirPath,
					"fstype":    "ext4",
					"run_fsck":  "no",
				},
			},
		)

		for i := range r.AllowedIPs {
			agents = append(agents, reactor.ResourceAgent{
				Type: "ocf:heartbeat:exportfs",
				Name: fmt.Sprintf(exportAgentName, vol.VolumeNumber, i),
				Attributes: map[string]string{
					"directory":                  dirPath,
					"fsid":                       fsid.String(),
					"clientspec":                 nfsFormatCidr(&r.AllowedIPs[i]),
					"options":                    "rw",
					"wait_for_leasetime_on_stop": "1",
				},
			})
		}
	}

	agents = append(agents, reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip", Attributes: map[string]string{"ip": r.ServiceIP.IP().String(), "cidr_netmask": strconv.Itoa(r.ServiceIP.Prefix())}})

	return &reactor.PromoterConfig{
		ID: r.ID(),
		Resources: map[string]reactor.PromoterResourceConfig{
			r.Name: {
				Runner:             "systemd",
				Start:              agents,
				StopServicesOnExit: true,
				OnStopFailure:      "echo b > /proc/sysrq-trigger",
				TargetAs:           "BindsTo",
			},
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
