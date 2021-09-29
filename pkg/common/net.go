package common

import (
	"encoding/json"
	"fmt"
	"net"
)

type IpCidr struct {
	net.IPNet
}

func (s *IpCidr) IP() net.IP {
	return s.IPNet.IP
}

func (s *IpCidr) Prefix() int {
	ones, _ := s.Mask.Size()
	return ones
}

func (s *IpCidr) Type() string {
	return "ip-cidr"
}

func (s *IpCidr) Set(raw string) error {
	service, err := ServiceIPFromString(raw)
	if err != nil {
		return err
	}

	*s = service
	return nil
}

func (s IpCidr) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.IPNet.String())
}

func (s *IpCidr) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	ser, err := ServiceIPFromString(str)
	if err != nil {
		return err
	}

	s.IPNet = ser.IPNet

	return nil
}

func ServiceIPFromString(s string) (IpCidr, error) {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return IpCidr{}, fmt.Errorf("failed to parse service ip: %w", err)
	}

	ipnet.IP = ip
	return IpCidr{IPNet: *ipnet}, nil
}

func ServiceIPFromParts(ip net.IP, prefix int) IpCidr {
	bits := 32
	if ip.To4() == nil {
		bits = 128
	}

	return IpCidr{
		IPNet: net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(prefix, bits),
		},
	}
}
