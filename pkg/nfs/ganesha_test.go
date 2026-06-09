package nfs

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LINBIT/linstor-gateway/pkg/common"
)

func cidr(t *testing.T, s string) common.IpCidr {
	t.Helper()
	c, err := common.ServiceIPFromString(s)
	require.NoError(t, err)
	return c
}

func TestAllowedIPsToClients(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []common.IpCidr
		want string
	}{
		{"v4 catch-all", []common.IpCidr{common.ServiceIPFromParts(net.IPv4zero, 0)}, "*"},
		{"v6 catch-all", []common.IpCidr{common.ServiceIPFromParts(net.IPv6zero, 0)}, "*"},
		{"v4 network", []common.IpCidr{cidr(t, "10.20.0.0/16")}, "10.20.0.0/16"},
		{"v4 host /32", []common.IpCidr{cidr(t, "192.168.1.5/32")}, "192.168.1.5/32"},
		{"v6 host /128 -> bare", []common.IpCidr{cidr(t, "fd00::3/128")}, "fd00::3"},
		{"v6 network", []common.IpCidr{cidr(t, "fd00::/64")}, "fd00::/64"},
		{"multiple", []common.IpCidr{cidr(t, "10.0.0.0/8"), cidr(t, "192.168.0.0/16")}, "10.0.0.0/8,192.168.0.0/16"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, allowedIPsToClients(tc.in))
		})
	}
}

func TestClientsToAllowedIPs(t *testing.T) {
	t.Parallel()
	v4Service := common.ServiceIPFromParts(net.IP{192, 168, 0, 1}, 24)
	v6Service := common.ServiceIPFromParts(net.ParseIP("fd00::1"), 64)
	tests := []struct {
		name      string
		clients   string
		serviceIP common.IpCidr
		want      []string
	}{
		{"star v4 service", "*", v4Service, []string{"0.0.0.0/0"}},
		{"star v6 service", "*", v6Service, []string{"::/0"}},
		{"v4 network", "10.20.0.0/16", v4Service, []string{"10.20.0.0/16"}},
		{"bare v4 host", "192.168.1.5", v4Service, []string{"192.168.1.5/32"}},
		{"bare v6 host", "fd00::3", v6Service, []string{"fd00::3/128"}},
		{"multiple with spaces", "10.0.0.0/8, 192.168.0.0/16", v4Service, []string{"10.0.0.0/8", "192.168.0.0/16"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clientsToAllowedIPs(tc.clients, tc.serviceIP)
			gotStr := make([]string, len(got))
			for i := range got {
				gotStr[i] = got[i].String()
			}
			assert.Equal(t, tc.want, gotStr)
		})
	}
}

func TestGaneshaAgent(t *testing.T) {
	t.Parallel()
	service := common.ServiceIPFromParts(net.IP{192, 168, 0, 1}, 24)
	allowed := []common.IpCidr{cidr(t, "10.20.0.0/16")}

	t.Run("single volume", func(t *testing.T) {
		agent, err := ganeshaAgent(service,
			[]ganeshaExport{{path: "/srv/gateway-exports/nfs1/", id: 1}},
			allowed)
		assert.NoError(t, err)
		assert.Equal(t, "ocf:heartbeat:ganesha-nfs", agent.Type)
		assert.Equal(t, "nfsserver", agent.Name)
		assert.Equal(t, "192.168.0.1", agent.Attributes["nfs_ip"])
		assert.Equal(t, "/srv/gateway-exports/nfs1/", agent.Attributes["export_path"])
		assert.Equal(t, "1", agent.Attributes["export_id"])
		assert.Equal(t, "10.20.0.0/16", agent.Attributes["clients"])
		assert.Equal(t, "All_Squash", agent.Attributes["squash"])
	})

	t.Run("multiple volumes produce parallel ;-lists", func(t *testing.T) {
		agent, err := ganeshaAgent(service,
			[]ganeshaExport{
				{path: "/srv/gateway-exports/nfs1/music", id: 1},
				{path: "/srv/gateway-exports/nfs1/movies", id: 2},
			},
			allowed)
		assert.NoError(t, err)
		assert.Equal(t, "/srv/gateway-exports/nfs1/music;/srv/gateway-exports/nfs1/movies", agent.Attributes["export_path"])
		assert.Equal(t, "1;2", agent.Attributes["export_id"])
	})

	t.Run("no allowed IPs is an error (deny-default)", func(t *testing.T) {
		_, err := ganeshaAgent(service,
			[]ganeshaExport{{path: "/srv/gateway-exports/nfs1/", id: 1}}, nil)
		assert.Error(t, err)
	})

	t.Run("no exports is an error", func(t *testing.T) {
		_, err := ganeshaAgent(service, nil, allowed)
		assert.Error(t, err)
	})
}

func TestClientsRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		serviceIP common.IpCidr
		orig      []common.IpCidr
	}{
		{
			name:      "v4",
			serviceIP: common.ServiceIPFromParts(net.IP{192, 168, 0, 1}, 24),
			orig:      []common.IpCidr{cidr(t, "10.20.0.0/16"), cidr(t, "192.168.1.5/32")},
		},
		{
			// The /128 -> bare -> /128 path is the trickiest part of the
			// encoding, so exercise it end-to-end alongside a v6 network.
			name:      "v6",
			serviceIP: common.ServiceIPFromParts(net.ParseIP("fd00::1"), 64),
			orig:      []common.IpCidr{cidr(t, "fd00::/64"), cidr(t, "fd00::3/128")},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			back := clientsToAllowedIPs(allowedIPsToClients(tc.orig), tc.serviceIP)
			assert.Len(t, back, len(tc.orig))
			for i := range tc.orig {
				assert.Equal(t, tc.orig[i].String(), back[i].String())
			}
		})
	}
}
