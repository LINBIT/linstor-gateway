package nvmeof

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor/client"
	"github.com/google/uuid"
	"github.com/icza/gog"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

const IDFormat = "nvmeof-%s"
const DefaultPort = 4420

type ResourceConfig struct {
	NQN           Nqn                   `json:"nqn"`
	ServiceIP     common.IpCidr         `json:"service_ip"`
	ResourceGroup string                `json:"resource_group"`
	Volumes       []common.VolumeConfig `json:"volumes"`
	Status        common.ResourceStatus `json:"status"`
	GrossSize     bool                  `json:"gross_size"`
}

func (r *ResourceConfig) VolumeConfig(number int) *common.Volume {
	var result *common.Volume

	for _, vol := range r.Volumes {
		if vol.Number == number {
			result = &common.Volume{
				Volume: vol,
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

func (r *ResourceConfig) ID() string {
	return fmt.Sprintf(IDFormat, r.NQN.Subsystem())
}

func parseIP(startEntries []reactor.StartEntry, index int) (common.IpCidr, error) {
	ipAgent, ok := startEntries[index].(*reactor.ResourceAgent)
	if !ok {
		return common.IpCidr{}, fmt.Errorf("expected a resource agent at index %d, got a systemd service", index)
	}
	if ipAgent.Type != "ocf:heartbeat:IPaddr2" {
		return common.IpCidr{}, fmt.Errorf("expected 'ocf:heartbeat:IPaddr2' agent, got '%s' instead", ipAgent.Type)
	}

	ip := net.ParseIP(ipAgent.Attributes["ip"])
	if ip == nil {
		return common.IpCidr{}, fmt.Errorf("malformed ip %s", ipAgent.Attributes["ip"])
	}

	prefixLength, err := strconv.Atoi(ipAgent.Attributes["cidr_netmask"])
	if err != nil {
		return common.IpCidr{}, fmt.Errorf("failed to parse service ip prefix")
	}

	return common.ServiceIPFromParts(ip, prefixLength), nil
}

func parseNQN(startEntries []reactor.StartEntry, index int) (Nqn, error) {
	subsysAgent, ok := startEntries[index].(*reactor.ResourceAgent)
	if !ok {
		return Nqn{}, fmt.Errorf("expected a resource agent at index %d, got a systemd service", index)
	}
	if subsysAgent.Type != "ocf:heartbeat:nvmet-subsystem" {
		return Nqn{}, errors.New(fmt.Sprintf("expected 'ocf:heartbeat:nvmet-subsystem' agent, got '%s' instead", subsysAgent.Type))
	}

	return NewNqn(subsysAgent.Attributes["nqn"])
}

func FromPromoter(cfg *reactor.PromoterConfig, definition *client.ResourceDefinition, volumeDefinition []client.VolumeDefinition) (*ResourceConfig, error) {
	r := &ResourceConfig{}
	var nqn string
	n, err := fmt.Sscanf(cfg.ID, IDFormat, &nqn)
	if n != 1 {
		return nil, fmt.Errorf("failed to parse id into resource name: %w", err)
	}

	r.ResourceGroup = definition.ResourceGroupName

	if len(cfg.Resources) != 1 {
		return nil, errors.New(fmt.Sprintf("promoter config without exactly 1 resource (has %d)", len(cfg.Resources)))
	}

	var rscCfg reactor.PromoterResourceConfig
	for _, v := range cfg.Resources {
		rscCfg = v
	}

	if len(rscCfg.Start) < 5 {
		return nil, errors.New(fmt.Sprintf("config has too few agent entries, expected at least 3, got %d", len(rscCfg.Start)))
	}

	r.ServiceIP, err = parseIP(rscCfg.Start, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service IP: %w", err)
	}

	r.NQN, err = parseNQN(rscCfg.Start, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NQN: %w", err)
	}
	for _, vd := range volumeDefinition {
		if vd.VolumeNumber == nil {
			vd.VolumeNumber = gog.Ptr(int32(0))
		}
		r.Volumes = append(r.Volumes, common.VolumeConfig{
			Number:  int(*vd.VolumeNumber),
			SizeKiB: vd.SizeKib,
		})
	}

	return r, nil
}

func (r *ResourceConfig) ToPromoter(deployment []client.ResourceWithVolumes) (*reactor.PromoterConfig, error) {
	if len(deployment) == 0 {
		return nil, errors.New("resource config is missing deployment information")
	}
	deployedRes := deployment[0]

	digest := sha256.Sum256([]byte(r.NQN.String()))
	serial := hex.EncodeToString(digest[:8])

	uuidNS := uuid.NewSHA1(UUIDNVMeoF, []byte(deployedRes.Uuid))

	// volume 0 is reserved as the "cluster private" volume
	deployedClusterPrivateVol := deployedRes.Volumes[0]

	agents := []reactor.StartEntry{
		&reactor.ResourceAgent{
			Type: "ocf:heartbeat:portblock",
			Name: fmt.Sprintf("portblock"),
			Attributes: map[string]string{
				"ip":       r.ServiceIP.IP().String(),
				"portno":   strconv.Itoa(DefaultPort),
				"action":   "block",
				"protocol": "tcp",
			},
		},
		common.ClusterPrivateVolumeAgent(deployedClusterPrivateVol, r.NQN.Subsystem()),
		&reactor.ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "service_ip",
			Attributes: map[string]string{
				"ip":           r.ServiceIP.IP().String(),
				"cidr_netmask": strconv.Itoa(r.ServiceIP.Prefix()),
			},
		},
		&reactor.ResourceAgent{
			Type: "ocf:heartbeat:nvmet-subsystem",
			Name: "subsys",
			Attributes: map[string]string{
				"nqn":    r.NQN.String(),
				"serial": serial,
			},
		},
	}

	for i := 1; i < len(deployedRes.Volumes); i++ {
		vol := deployedRes.Volumes[i]
		if int(vol.VolumeNumber) != r.Volumes[i].Number {
			return nil, fmt.Errorf("inconsistent volumes, expected volume number %d, got %d", vol.VolumeNumber, r.Volumes[i].Number)
		}

		guid := uuid.NewSHA1(uuidNS, []byte(vol.Uuid))

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

		agents = append(agents, &reactor.ResourceAgent{
			Type: "ocf:heartbeat:nvmet-namespace",
			Name: fmt.Sprintf("ns_%d", vol.VolumeNumber),
			Attributes: map[string]string{
				"nqn": r.NQN.String(),
				// nvme namespaces start at 1, we just ensure that our volumes also start at 1
				"namespace_id": fmt.Sprintf("%d", vol.VolumeNumber),
				"backing_path": fmt.Sprintf(devPath),
				"uuid":         guid.String(),
				"nguid":        guid.String(),
			},
		})
	}

	agents = append(agents, &reactor.ResourceAgent{Type: "ocf:heartbeat:nvmet-port", Name: "port", Attributes: map[string]string{"nqns": r.NQN.String(), "addr": r.ServiceIP.IP().String(), "type": "tcp"}})

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:portblock",
		Name: "portunblock",
		Attributes: map[string]string{
			"ip":         r.ServiceIP.IP().String(),
			"portno":     strconv.Itoa(DefaultPort),
			"action":     "unblock",
			"protocol":   "tcp",
			"tickle_dir": filepath.Join(common.ClusterPrivateVolumeMountPath, deployedRes.Name),
		},
	})

	return &reactor.PromoterConfig{
		ID: r.ID(),
		Resources: map[string]reactor.PromoterResourceConfig{
			r.NQN.Subsystem(): {
				Runner:              "systemd",
				Start:               agents,
				StopServicesOnExit:  true,
				OnDrbdDemoteFailure: "reboot-immediate",
				TargetAs:            "Requires",
			},
		},
	}, nil
}

func (r *ResourceConfig) Matches(o *ResourceConfig) bool {
	if r.NQN != o.NQN {
		return false
	}

	if r.ServiceIP.String() != o.ServiceIP.String() {
		return false
	}

	if r.ResourceGroup != o.ResourceGroup {
		return false
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
	}

	return true
}

func (r *ResourceConfig) FillDefaults() {
	if r.ResourceGroup == "" {
		r.ResourceGroup = "DfltRscGrp"
	}
}

func (r *ResourceConfig) Valid() error {
	if len(r.NQN.Subsystem()) < 2 {
		return common.ValidationError("nvme subsystem string to short (min. 2)")
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

	for i := range r.Volumes {
		if i != 0 && r.Volumes[i].Number <= 0 {
			// the "cluster private volume" (volume 0) is excluded from this check
			return common.ValidationError("volume numbers must start at 1")
		}

		if r.Volumes[i].SizeKiB <= 0 {
			return common.ValidationError("volume size must be positive")
		}

		if i > 0 && r.Volumes[i-1].Number == r.Volumes[i].Number {
			return common.ValidationError("volume numbers must be unique")
		}
	}

	return nil
}
