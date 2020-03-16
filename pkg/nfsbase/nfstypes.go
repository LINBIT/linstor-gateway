package nfsbase

import "net"

type NfsConfig struct {
	ResourceName      string `json:"resource,omitempty"`
	ServiceIP         net.IP `json:"service_ip,omitempty"`
	ServiceIPNetBits  int    `json:"service_ip_mask,omitempty"`
	AllowedIPs        net.IP `json:"allowed_ips,omitempty"`
	AllowedIPsNetBits int    `json:"allowed_ips_mask,omitempty"`
	SizeKiB           uint64 `json:"size_kib,omitempty"`
}

