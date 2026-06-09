package nfs

import (
	"net"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/linstor-gateway/pkg/common"
)

// allowedIPsToClients renders a list of allowed CIDRs into the ganesha-nfs
// agent's "clients" parameter (a single comma-separated whitelist, shared by
// all exports). The catch-all (0.0.0.0/0 or ::/0) becomes "*"; an IPv6 host
// is emitted bare because ganesha's Clients directive rejects a /128 suffix;
// everything else is emitted as a CIDR.
func allowedIPsToClients(ips []common.IpCidr) string {
	parts := make([]string, 0, len(ips))
	for _, allowed := range ips {
		ip := allowed.IP()
		switch {
		case allowed.Prefix() == 0 && ip.IsUnspecified():
			parts = append(parts, "*")
		case ip.To4() == nil && allowed.Prefix() == 128:
			parts = append(parts, ip.String())
		default:
			parts = append(parts, allowed.String())
		}
	}
	return strings.Join(parts, ",")
}

// clientsToAllowedIPs is the inverse of allowedIPsToClients. "*" maps to the
// catch-all that matches the service IP's address family, so the value
// round-trips unambiguously. Bare host addresses are expanded to /32 (IPv4)
// or /128 (IPv6). Unparseable entries are logged and skipped.
func clientsToAllowedIPs(clients string, serviceIP common.IpCidr) []common.IpCidr {
	var out []common.IpCidr
	for _, raw := range strings.Split(clients, ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		if entry == "*" {
			if serviceIP.IP().To4() == nil {
				out = append(out, common.ServiceIPFromParts(net.IPv6zero, 0))
			} else {
				out = append(out, common.ServiceIPFromParts(net.IPv4zero, 0))
			}
			continue
		}
		if strings.Contains(entry, "/") {
			c, err := common.ServiceIPFromString(entry)
			if err != nil {
				log.Warnf("ignoring unparseable ganesha client %q: %v", entry, err)
				continue
			}
			out = append(out, c)
			continue
		}
		ip := net.ParseIP(entry)
		if ip == nil {
			log.Warnf("ignoring unparseable ganesha client %q", entry)
			continue
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		out = append(out, common.ServiceIPFromParts(ip, bits))
	}
	return out
}
