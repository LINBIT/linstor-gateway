package reactor

// SystemdService is an entry within a drbd-reactor config that describes a
// systemd service.
// It is very simple: the whole text is just the systemd service name.
type SystemdService struct {
	Name string
}

func (s *SystemdService) MarshalText() (text []byte, err error) {
	return []byte(s.Name), nil
}

func (s *SystemdService) UnmarshalText(text []byte) error {
	s.Name = string(text)
	return nil
}
