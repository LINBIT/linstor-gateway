package reactor

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"bitbucket.org/creachadair/shell"
)

// ResourceAgent is an entry within a drbd-reactor config that describes an
// ocf resource agent.
// The structure of such an ocf resource definition is as follows:
//
//     ocf:$vendor:$agent $instance-id name=value ...
//
// For details on how these, fields are decoded, see UnmarshalText.
type ResourceAgent struct {
	Type       string
	Name       string
	Attributes map[string]string
}

// UnmarshalText parses a ResourceAgent from its string representation, as
// defined by the drbd-reactor configuration format.
// The structure of such an ocf resource definition is as follows:
//
//     ocf:$vendor:$agent $instance-id name=value ...
//
// The first part, "ocf:$vendor:$agent" will be put into the "Type" field of the
// resulting ResourceAgent struct.
// $instance-id is the unique name of the resource agent instance, and will end
// up in the "Name" field of the ResourceAgent struct.
// After these fields follow an arbitrary number of optional key-value pairs.
// They will be parsed into the "Attributes" map of the ResourceAgent struct.
func (r *ResourceAgent) UnmarshalText(text []byte) error {
	parts, valid := shell.Split(string(text))
	if !valid || len(parts) < 2 {
		return errors.New("expected at least type and name")
	}

	r.Type = parts[0]
	r.Name = parts[1]
	for _, arg := range parts[2:] {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			return errors.New("expected key=value pairs as arguments")
		}

		if r.Attributes == nil {
			r.Attributes = map[string]string{}
		}
		r.Attributes[kv[0]] = kv[1]
	}

	return nil
}

func (r ResourceAgent) MarshalText() (text []byte, err error) {
	if r.Type == "" {
		return nil, fmt.Errorf("invalid resource agent without type")
	}
	if r.Name == "" {
		return nil, fmt.Errorf("invalid resource agent without name")
	}
	args := make([]string, 0, len(r.Attributes))
	for k, v := range r.Attributes {
		args = append(args, fmt.Sprintf("%s=%s", k, shell.Quote(v)))
	}

	// Ensure consistent serialization order
	sort.Strings(args)

	return []byte(strings.Trim(fmt.Sprintf("%s %s %s", r.Type, r.Name, strings.Join(args, " ")), " ")), nil
}

var _ StartEntry = &ResourceAgent{}
