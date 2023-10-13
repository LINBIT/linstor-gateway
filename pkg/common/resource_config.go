package common

import (
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
	log "github.com/sirupsen/logrus"
	"net"
	"path/filepath"
	"strings"
)

const (
	clusterPrivateVolumeSizeKiB    = 64 * 1024 // 64MiB
	clusterPrivateVolumeFileSystem = "ext4"
	ClusterPrivateVolumeMountPath  = "/srv/ha/internal"
	ClusterPrivateVolumeAgentName  = "fs_cluster_private"
)

func DevicePath(vol client.Volume) string {
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
	return devPath
}

func ClusterPrivateVolume() VolumeConfig {
	return VolumeConfig{
		Number:              0,
		SizeKiB:             clusterPrivateVolumeSizeKiB,
		FileSystem:          clusterPrivateVolumeFileSystem,
		FileSystemRootOwner: UidGid{Uid: 0, Gid: 0},
	}
}

func ClusterPrivateVolumeAgent(deployedVol client.Volume, resource string) *reactor.ResourceAgent {
	return &reactor.ResourceAgent{
		Type: "ocf:heartbeat:Filesystem",
		Name: ClusterPrivateVolumeAgentName,
		Attributes: map[string]string{
			"device":    DevicePath(deployedVol),
			"directory": filepath.Join(ClusterPrivateVolumeMountPath, resource),
			"fstype":    clusterPrivateVolumeFileSystem,
			"run_fsck":  "no",
		},
	}
}

func CheckIPCollision(config reactor.PromoterConfig, checkIP net.IP) error {
	name, rscCfg := config.FirstResource()
	if rscCfg == nil {
		return fmt.Errorf("no resource found in config")
	}
	for _, entry := range rscCfg.Start {
		switch agent := entry.(type) {
		case *reactor.ResourceAgent:
			switch agent.Type {
			case "ocf:heartbeat:IPaddr2":
				ip := net.ParseIP(agent.Attributes["ip"])
				if ip == nil {
					return fmt.Errorf("malformed IP address %s in agent %s of config %s",
						agent.Attributes["ip"], agent.Name, name)
				}
				log.Debugf("checking IP %s", ip)
				if ip.Equal(checkIP) {
					return fmt.Errorf("IP address %s already in use by config %s",
						ip.String(), name)
				}
			}
		}
	}
	return nil
}
