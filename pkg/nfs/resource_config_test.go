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
			ClusterPrivateVolume("test"),
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
