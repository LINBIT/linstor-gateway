package nvmeof

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Nqn represents a conventional nvme qualified name
type Nqn [2]string

func NewNqn(s string) (Nqn, error) {
	n := Nqn{}
	err := n.UnmarshalText([]byte(s))
	if err != nil {
		return Nqn{}, err
	}

	return n, nil
}

func (n *Nqn) Vendor() string {
	return n[0]
}

func (n *Nqn) Subsystem() string {
	return n[1]
}

func (n *Nqn) UnmarshalText(text []byte) error {
	s := string(text)
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return malformedNqn(s)
	}

	n[0] = parts[0]
	n[1] = parts[2]
	return nil
}

func (n *Nqn) UnmarshalJSON(text []byte) error {
	var s string
	err := json.Unmarshal(text, &s)
	if err != nil {
		return err
	}
	return n.UnmarshalText([]byte(s))
}

func (n Nqn) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n Nqn) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n Nqn) String() string {
	return fmt.Sprintf("%s:nvme:%s", n[0], n[1])
}

type malformedNqn string

func (m malformedNqn) Error() string {
	return fmt.Sprintf("NQN '%s' malformed, expected <vendor>:nvme:<subsystem>", string(m))
}
