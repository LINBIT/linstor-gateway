package iscsi

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/LINBIT/golinstor/client"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

const (
	DefaultISCSIPort = 3260
)

type ResourceConfig struct {
	IQN               Iqn                   `json:"iqn"`
	AllowedInitiators []Iqn                 `json:"allowed_initiators,omitempty"`
	ResourceGroup     string                `json:"resource_group"`
	Volumes           []common.VolumeConfig `json:"volumes"`
	Username          string                `json:"username,omitempty"`
	Password          string                `json:"password,omitempty"`
	ServiceIPs        []common.IpCidr       `json:"service_ips"`
	Status            common.ResourceStatus `json:"status"`
}

const (
	agentTypePortblock   = "ocf:heartbeat:portblock"
	agentTypeIPaddr2     = "ocf:heartbeat:IPaddr2"
	agentTypeISCSITarget = "ocf:heartbeat:iSCSITarget"
)

const minAgentEntries = 4 // portblock, service_ip, target, portunblock

func parsePromoterConfig(cfg *reactor.PromoterConfig) (*ResourceConfig, error) {
	r := &ResourceConfig{}
	var res string
	n, err := fmt.Sscanf(cfg.ID, IDFormat, &res)
	if n != 1 {
		return nil, fmt.Errorf("failed to parse id into resource name: %w", err)
	}

	if len(cfg.Resources) != 1 {
		return nil, errors.New(fmt.Sprintf("promoter config without exactly 1 resource (has %d)", len(cfg.Resources)))
	}

	var rscCfg reactor.PromoterResourceConfig
	for _, v := range cfg.Resources {
		rscCfg = v
	}

	if len(rscCfg.Start) < minAgentEntries {
		return nil, errors.New(fmt.Sprintf("config has too few agent entries, expected at least %d, got %d",
			minAgentEntries, len(rscCfg.Start)))
	}

	var numPortblocks, numPortunblocks int
	for _, entry := range rscCfg.Start {
		switch agent := entry.(type) {
		case *reactor.ResourceAgent:
			switch agent.Type {
			case agentTypePortblock:
				switch agent.Attributes["action"] {
				case "block":
					numPortblocks++
				case "unblock":
					numPortunblocks++
				}
			case agentTypeIPaddr2:
				ip := net.ParseIP(agent.Attributes["ip"])
				if ip == nil {
					return nil, fmt.Errorf("malformed ip %s", agent.Attributes["ip"])
				}

				prefixLength, err := strconv.Atoi(agent.Attributes["cidr_netmask"])
				if err != nil {
					return nil, fmt.Errorf("failed to parse service ip prefix")
				}

				r.ServiceIPs = append(r.ServiceIPs, common.ServiceIPFromParts(ip, prefixLength))
			case agentTypeISCSITarget:
				r.IQN, err = NewIqn(agent.Attributes["iqn"])
				if err != nil {
					return nil, fmt.Errorf("got malformed iqn: %w", err)
				}

				r.Username = agent.Attributes["incoming_username"]
				r.Password = agent.Attributes["incoming_password"]

				rawAllowed := agent.Attributes["allowed_initiators"]
				if rawAllowed != "" {
					for _, allowed := range strings.Split(rawAllowed, " ") {
						iqn, err := NewIqn(allowed)
						if err != nil {
							return nil, fmt.Errorf("got malformed iqn %s for allowed initiators: %w", allowed, err)
						}
						r.AllowedInitiators = append(r.AllowedInitiators, iqn)
					}
				}
			}
		case *reactor.SystemdService:
			// ignore systemd services for now
		}
	}

	if numPortblocks != numPortunblocks {
		return nil, fmt.Errorf("malformed configuration: got a different number of portblock and portunblock agents")
	}

	if numPortblocks != len(r.ServiceIPs) {
		return nil, fmt.Errorf("malformed configuration: got a different number of portblock agents than IPaddr2 agents")
	}

	return r, nil
}

func FromPromoter(cfg *reactor.PromoterConfig, definition *client.ResourceDefinition, volumeDefinitions []client.VolumeDefinition) (*ResourceConfig, error) {
	r, err := parsePromoterConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse promoter config: %w", err)
	}
	r.ResourceGroup = definition.ResourceGroupName

	for _, vd := range volumeDefinitions {
		r.Volumes = append(r.Volumes, common.VolumeConfig{
			Number:  int(vd.VolumeNumber),
			SizeKiB: vd.SizeKib,
		})
	}

	return r, nil
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

func (r *ResourceConfig) FillDefaults() {
	if r.ResourceGroup == "" {
		r.ResourceGroup = "DfltRscGrp"
	}

	for i := range r.Volumes {
		if r.Volumes[i].Number == 0 {
			r.Volumes[i].Number = i + 1
		}
	}
}

func (r *ResourceConfig) Valid() error {
	if len(r.IQN.WWN()) < 2 {
		return common.ValidationError("iscsi wwn string to short (min. 2)")
	}

	if len(r.ServiceIPs) == 0 {
		return common.ValidationError("missing service ips")
	}

	sort.Slice(r.Volumes, func(i, j int) bool {
		return r.Volumes[i].Number < r.Volumes[j].Number
	})

	for i := range r.Volumes {
		if r.Volumes[i].Number < 1 {
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

func (r *ResourceConfig) Matches(o *ResourceConfig) bool {
	if r.IQN != o.IQN {
		return false
	}

	for i := range r.ServiceIPs {
		if r.ServiceIPs[i].String() != o.ServiceIPs[i].String() {
			return false
		}
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

	if r.Username != o.Username {
		return false
	}

	if r.Password != o.Password {
		return false
	}

	return true
}

func (r *ResourceConfig) portals() string {
	var portals []string
	for _, ip := range r.ServiceIPs {
		portals = append(portals, fmt.Sprintf("%s:%d", ip.IP(), DefaultISCSIPort))
	}
	return strings.Join(portals, " ")
}

func (r *ResourceConfig) ID() string {
	return fmt.Sprintf(IDFormat, r.IQN.WWN())
}

func (r *ResourceConfig) ToPromoter(deployment []client.ResourceWithVolumes) (*reactor.PromoterConfig, error) {
	if len(deployment) == 0 {
		return nil, errors.New("resource config is missing deployment information")
	}

	allowedInitiatorStrings := make([]string, 0, len(r.AllowedInitiators))
	for i := range r.AllowedInitiators {
		allowedInitiatorStrings = append(allowedInitiatorStrings, r.AllowedInitiators[i].String())
	}

	var agents []reactor.StartEntry
	for i, ip := range r.ServiceIPs {
		agents = append(agents, &reactor.ResourceAgent{
			Type: "ocf:heartbeat:portblock",
			Name: fmt.Sprintf("pblock%d", i),
			Attributes: map[string]string{
				"ip":       ip.IP().String(),
				"portno":   strconv.Itoa(DefaultISCSIPort),
				"action":   "block",
				"protocol": "tcp",
			},
		})
	}

	for i, ip := range r.ServiceIPs {
		agents = append(agents, &reactor.ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: fmt.Sprintf("service_ip%d", i),
			Attributes: map[string]string{
				"ip":           ip.IP().String(),
				"cidr_netmask": strconv.Itoa(ip.Prefix()),
			},
		})
	}

	agents = append(agents, &reactor.ResourceAgent{
		Type: "ocf:heartbeat:iSCSITarget",
		Name: "target",
		Attributes: map[string]string{
			"iqn":                r.IQN.String(),
			"portals":            r.portals(),
			"incoming_username":  r.Username,
			"incoming_password":  r.Password,
			"allowed_initiators": strings.Join(allowedInitiatorStrings, " "),
		},
	})

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

		agents = append(agents, &reactor.ResourceAgent{
			Type: "ocf:heartbeat:iSCSILogicalUnit",
			Name: fmt.Sprintf("lu%d", vol.VolumeNumber),
			Attributes: map[string]string{
				"target_iqn": r.IQN.String(),
				"lun":        strconv.Itoa(int(vol.VolumeNumber)),
				"path":       fmt.Sprintf(devPath),
				"product_id": "LINSTOR iSCSI",
			},
		})
	}

	for i, ip := range r.ServiceIPs {
		agents = append(agents, &reactor.ResourceAgent{
			Type: "ocf:heartbeat:portblock",
			Name: fmt.Sprintf("portunblock%d", i),
			Attributes: map[string]string{
				"ip":       ip.IP().String(),
				"portno":   strconv.Itoa(DefaultISCSIPort),
				"action":   "unblock",
				"protocol": "tcp",
			},
		})
	}

	return &reactor.PromoterConfig{
		ID: r.ID(),
		Resources: map[string]reactor.PromoterResourceConfig{
			r.IQN.WWN(): {
				Runner:              "systemd",
				Start:               agents,
				StopServicesOnExit:  true,
				OnDrbdDemoteFailure: "reboot-immediate",
				TargetAs:            "Requires",
			},
		},
	}, nil
}
