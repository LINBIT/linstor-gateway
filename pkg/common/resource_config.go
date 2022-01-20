package common

import (
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
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

func ClusterPrivateVolumeAgent(vol VolumeConfig, deployedVol client.Volume, resource string) *reactor.ResourceAgent {
	return &reactor.ResourceAgent{
		Type: "ocf:heartbeat:Filesystem",
		Name: ClusterPrivateVolumeAgentName,
		Attributes: map[string]string{
			"device":    DevicePath(deployedVol),
			"directory": filepath.Join(ClusterPrivateVolumeMountPath, resource),
			"fstype":    vol.FileSystem,
			"run_fsck":  "no",
		},
	}
}
