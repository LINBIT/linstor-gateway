package nvmeof_test

import (
	"github.com/icza/gog"
	"net"
	"testing"

	"github.com/LINBIT/golinstor/client"
	"github.com/stretchr/testify/assert"

	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/nvmeof"
)

func TestResource_RoundTrip(t *testing.T) {
	t.Parallel()

	testcases := []nvmeof.ResourceConfig{
		{
			NQN: nvmeof.Nqn{"nqn.com.example.test", "example-resource"},
			Volumes: []common.VolumeConfig{
				{Number: 2, SizeKiB: 1024},
			},
			ResourceGroup: "rg1",
			ServiceIP:     common.ServiceIPFromParts(net.IP{192, 168, 127, 1}, 24),
		},
	}

	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.NQN.String(), func(t *testing.T) {
			t.Parallel()

			encoded, err := tcase.ToPromoter([]client.ResourceWithVolumes{
				{Volumes: []client.Volume{{VolumeNumber: 2, DevicePath: "/dev/drbd1002"}}},
			})
			assert.NoError(t, err)

			decoded, err := nvmeof.FromPromoter(
				encoded,
				&client.ResourceDefinition{ResourceGroupName: "rg1"},
				[]client.VolumeDefinition{
					{VolumeNumber: gog.Ptr(int32(2)), SizeKib: 1024},
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, tcase.NQN, decoded.NQN)
			assert.Equal(t, tcase.ServiceIP.String(), decoded.ServiceIP.String())
			assert.Equal(t, tcase.Volumes, decoded.Volumes)
			assert.Equal(t, tcase.ResourceGroup, decoded.ResourceGroup)
		})
	}
}
