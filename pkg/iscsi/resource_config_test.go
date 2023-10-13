package iscsi

import (
	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ipnet(str string) common.IpCidr {
	ip, err := common.ServiceIPFromString(str)
	if err != nil {
		panic(err)
	}
	return ip
}

func TestParsePromoterConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *reactor.PromoterConfig
		want    *ResourceConfig
		wantErr bool
	}{
		{
			name: "default",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock0", Attributes: map[string]string{"action": "block", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip0", Attributes: map[string]string{"cidr_netmask": "16", "ip": "1.1.1.1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSITarget", Name: "target", Attributes: map[string]string{"allowed_initiators": "", "incoming_username": "user", "incoming_password": "password", "iqn": "iqn.2021-08.com.linbit:target1", "portals": "1.1.1.1:3260 2.2.2.2:3260 3.3.3.3:3260"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSILogicalUnit", Name: "lu1", Attributes: map[string]string{"lun": "1", "path": "/dev/drbd/by-res/target1/1", "product_id": "LINSTOR iSCSI", "target_iqn": "iqn.2021-08.com.linbit:target1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock0", Attributes: map[string]string{"action": "unblock", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
						},
					},
				},
			},
			want: &ResourceConfig{
				IQN: Iqn{"iqn.2021-08.com.linbit", "target1"}, AllowedInitiators: nil, Username: "user", Password: "password",
				ServiceIPs: []common.IpCidr{ipnet("1.1.1.1/16")},
			},
		},
		{
			name: "invalid id",
			cfg: &reactor.PromoterConfig{
				ID: "target1",
			},
			wantErr: true,
		},
		{
			name: "too many resources",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {},
					"target2": {},
				},
			},
			wantErr: true,
		},
		{
			name: "too few agent entries",
			cfg: &reactor.PromoterConfig{
				ID:        "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{"target1": {Start: []reactor.StartEntry{}}},
			},
			wantErr: true,
		},
		{
			name: "missing portblock",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip0", Attributes: map[string]string{"cidr_netmask": "16", "ip": "1.1.1.1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSITarget", Name: "target", Attributes: map[string]string{"allowed_initiators": "", "incoming_username": "user", "incoming_password": "password", "iqn": "iqn.2021-08.com.linbit:target1", "portals": "1.1.1.1:3260 2.2.2.2:3260 3.3.3.3:3260"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSILogicalUnit", Name: "lu1", Attributes: map[string]string{"lun": "1", "path": "/dev/drbd/by-res/target1/1", "product_id": "LINSTOR iSCSI", "target_iqn": "iqn.2021-08.com.linbit:target1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock0", Attributes: map[string]string{"action": "unblock", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing ip",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock0", Attributes: map[string]string{"action": "block", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSITarget", Name: "target", Attributes: map[string]string{"allowed_initiators": "", "incoming_username": "user", "incoming_password": "password", "iqn": "iqn.2021-08.com.linbit:target1", "portals": "1.1.1.1:3260 2.2.2.2:3260 3.3.3.3:3260"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSILogicalUnit", Name: "lu1", Attributes: map[string]string{"lun": "1", "path": "/dev/drbd/by-res/target1/1", "product_id": "LINSTOR iSCSI", "target_iqn": "iqn.2021-08.com.linbit:target1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock0", Attributes: map[string]string{"action": "unblock", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple ips",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock0", Attributes: map[string]string{"action": "block", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock1", Attributes: map[string]string{"action": "block", "ip": "2.2.2.2", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock2", Attributes: map[string]string{"action": "block", "ip": "3.3.3.3", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip0", Attributes: map[string]string{"cidr_netmask": "16", "ip": "1.1.1.1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip1", Attributes: map[string]string{"cidr_netmask": "16", "ip": "2.2.2.2"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip2", Attributes: map[string]string{"cidr_netmask": "16", "ip": "3.3.3.3"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSITarget", Name: "target", Attributes: map[string]string{"allowed_initiators": "", "incoming_username": "user", "incoming_password": "password", "iqn": "iqn.2021-08.com.linbit:target1", "portals": "1.1.1.1:3260 2.2.2.2:3260 3.3.3.3:3260"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSILogicalUnit", Name: "lu1", Attributes: map[string]string{"lun": "1", "path": "/dev/drbd/by-res/target1/1", "product_id": "LINSTOR iSCSI", "target_iqn": "iqn.2021-08.com.linbit:target1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock0", Attributes: map[string]string{"action": "unblock", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock1", Attributes: map[string]string{"action": "unblock", "ip": "2.2.2.2", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock2", Attributes: map[string]string{"action": "unblock", "ip": "3.3.3.3", "portno": "3260", "protocol": "tcp"}},
						},
					},
				},
			},
			want: &ResourceConfig{
				IQN: Iqn{"iqn.2021-08.com.linbit", "target1"}, AllowedInitiators: nil, Username: "user", Password: "password",
				ServiceIPs: []common.IpCidr{ipnet("1.1.1.1/16"), ipnet("2.2.2.2/16"), ipnet("3.3.3.3/16")},
			},
		},
		{
			name: "with specific implementation",
			cfg: &reactor.PromoterConfig{
				ID: "iscsi-target1",
				Resources: map[string]reactor.PromoterResourceConfig{
					"target1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "pblock0", Attributes: map[string]string{"action": "block", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:IPaddr2", Name: "service_ip0", Attributes: map[string]string{"cidr_netmask": "16", "ip": "1.1.1.1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSITarget", Name: "target", Attributes: map[string]string{"implementation": "scst", "allowed_initiators": "", "incoming_username": "user", "incoming_password": "password", "iqn": "iqn.2021-08.com.linbit:target1", "portals": "1.1.1.1:3260 2.2.2.2:3260 3.3.3.3:3260"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:iSCSILogicalUnit", Name: "lu1", Attributes: map[string]string{"implementation": "scst", "lun": "1", "path": "/dev/drbd/by-res/target1/1", "product_id": "LINSTOR iSCSI", "target_iqn": "iqn.2021-08.com.linbit:target1"}},
							&reactor.ResourceAgent{Type: "ocf:heartbeat:portblock", Name: "punblock0", Attributes: map[string]string{"action": "unblock", "ip": "1.1.1.1", "portno": "3260", "protocol": "tcp"}},
						},
					},
				},
			},
			want: &ResourceConfig{
				IQN: Iqn{"iqn.2021-08.com.linbit", "target1"}, AllowedInitiators: nil, Username: "user", Password: "password",
				ServiceIPs: []common.IpCidr{ipnet("1.1.1.1/16")}, Implementation: "scst",
			},
		},
	}
	for i := range tests {
		tcase := &tests[i]
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()
			got, err := parsePromoterConfig(tcase.cfg)
			if tcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tcase.want, got)
			}
		})
	}
}
