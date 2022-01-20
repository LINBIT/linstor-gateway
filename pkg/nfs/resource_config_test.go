package nfs

import (
	"github.com/BurntSushi/toml"
	apiconsts "github.com/LINBIT/golinstor"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
)

func TestResource_RoundTrip(t *testing.T) {
	t.Parallel()
	testcases := []ResourceConfig{{
		Name:      "test",
		ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
		AllowedIPs: []common.IpCidr{
			common.ServiceIPFromParts(net.IP{192, 168, 127, 0}, 24),
		},
		ResourceGroup: "rg1",
		Volumes: []VolumeConfig{
			{VolumeConfig: common.ClusterPrivateVolume()},
			{
				VolumeConfig: common.VolumeConfig{
					Number:     1,
					SizeKiB:    1024,
					FileSystem: "ext4",
				},
				ExportPath: "/",
			},
		},
		Status: common.ResourceStatus{},
	}}

	propsFilesystemExt4 := map[string]string{apiconsts.NamespcFilesystem + "/Type": "ext4"}
	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.Name, func(t *testing.T) {
			t.Parallel()

			encoded, err := tcase.ToPromoter([]client.ResourceWithVolumes{
				{Volumes: []client.Volume{
					{VolumeNumber: 0, DevicePath: "/dev/drbd1000", Props: propsFilesystemExt4},
					{VolumeNumber: 1, DevicePath: "/dev/drbd1001", Props: propsFilesystemExt4},
				}},
			})
			assert.NoError(t, err)

			err = toml.NewEncoder(log.Writer()).Encode(encoded)
			assert.NoError(t, err)

			decoded, err := FromPromoter(
				encoded,
				&client.ResourceDefinition{ResourceGroupName: "rg1"},
				[]client.VolumeDefinition{
					{VolumeNumber: 0, SizeKib: 64 * 1024, Props: propsFilesystemExt4},
					{VolumeNumber: 1, SizeKib: 1024, Props: propsFilesystemExt4},
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, tcase.Name, decoded.Name)
			assert.Equal(t, tcase.ServiceIP.String(), decoded.ServiceIP.String())
			assert.Len(t, decoded.AllowedIPs, len(tcase.AllowedIPs))
			for i := 0; i < len(decoded.AllowedIPs); i++ {
				assert.Equal(t, tcase.AllowedIPs[i].String(), decoded.AllowedIPs[i].String())
			}
			assert.Equal(t, tcase.ResourceGroup, decoded.ResourceGroup)
			assert.Equal(t, tcase.Volumes, decoded.Volumes)
			assert.Equal(t, tcase.Status, decoded.Status)
		})
	}
}

func TestFindFilesystemAgentVolume(t *testing.T) {
	t.Parallel()
	volumes := []client.VolumeDefinition{
		{VolumeNumber: 0, SizeKib: 1024},
		{VolumeNumber: 1, SizeKib: 65536},
	}

	tests := []struct {
		name    string
		volumes []client.VolumeDefinition
		agent   *reactor.ResourceAgent
		want    *client.VolumeDefinition
		wantErr bool
	}{{
		name:    "nil agent",
		volumes: volumes,
		agent:   nil,
		wantErr: true,
	}, {
		name:    "invalid agent type",
		volumes: volumes,
		agent: &reactor.ResourceAgent{
			Type: "not:Filesystem",
			Name: "fs_1",
		},
		wantErr: true,
	}, {
		name:    "normal agent",
		volumes: volumes,
		agent: &reactor.ResourceAgent{
			Type: "ocf:heartbeat:Filesystem",
			Name: "fs_1",
		},
		want: &volumes[1],
	}, {
		name:    "out of range volume number",
		volumes: volumes,
		agent: &reactor.ResourceAgent{
			Type: "ocf:heartbeat:Filesystem",
			Name: "fs_27",
		},
		want: nil,
	}, {
		name:    "invalid name format",
		volumes: volumes,
		agent: &reactor.ResourceAgent{
			Type: "ocf:heartbeat:Filesystem",
			Name: "not_the_right_format",
		},
		wantErr: true,
	}, {
		name:    "cluster private volume",
		volumes: volumes,
		agent: &reactor.ResourceAgent{
			Type: "ocf:heartbeat:Filesystem",
			Name: "fs_cluster_private",
		},
		want: &volumes[0],
	}}
	for i := range tests {
		tcase := &tests[i]
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()
			got, err := findFilesystemAgentVolume(tcase.volumes, tcase.agent)
			if tcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tcase.want, got)
			}
		})
	}
}

func TestValid(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		config      ResourceConfig
		expectError bool
	}{{
		config: ResourceConfig{
			Name: "empty",
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "minimal_valid",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
		},
		expectError: false,
	}, {
		config: ResourceConfig{
			Name: "invalid_ip",
			ServiceIP: common.IpCidr{
				IPNet: net.IPNet{
					IP:   net.ParseIP("192.168.127.1"),
					Mask: nil,
				},
			},
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "x",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "invalid_volume_number",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
			Volumes: []VolumeConfig{
				{VolumeConfig: common.VolumeConfig{Number: -1}},
			},
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "invalid_size",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
			Volumes: []VolumeConfig{
				{VolumeConfig: common.VolumeConfig{Number: 1, SizeKiB: 0}},
			},
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "duplicate_volume_number",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
			Volumes: []VolumeConfig{
				{VolumeConfig: common.VolumeConfig{Number: 1, SizeKiB: 1024}},
				{VolumeConfig: common.VolumeConfig{Number: 1, SizeKiB: 1024}},
			},
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "duplicate_export_paths",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
			Volumes: []VolumeConfig{
				{VolumeConfig: common.VolumeConfig{Number: 1, SizeKiB: 1024}, ExportPath: "/xyz"},
				{VolumeConfig: common.VolumeConfig{Number: 2, SizeKiB: 1024}, ExportPath: "/xyz"},
			},
		},
		expectError: true,
	}, {
		config: ResourceConfig{
			Name:      "everything",
			ServiceIP: common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
			AllowedIPs: []common.IpCidr{
				common.ServiceIPFromParts(net.IP{192, 168, 127, 0}, 24),
			},
			ResourceGroup: "rg1",
			Volumes: []VolumeConfig{
				{VolumeConfig: common.ClusterPrivateVolume()},
				{
					VolumeConfig: common.VolumeConfig{
						Number:     1,
						SizeKiB:    1024,
						FileSystem: "ext4",
					},
					ExportPath: "/",
				},
			},
		},
		expectError: false,
	}}

	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.config.Name, func(t *testing.T) {
			t.Parallel()

			tcase.config.FillDefaults()
			err := tcase.config.Valid()
			if tcase.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
