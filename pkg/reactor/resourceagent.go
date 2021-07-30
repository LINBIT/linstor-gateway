package reactor

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type ResourceAgent struct {
	Type       string
	Name       string
	Attributes map[string]string
}

func (r *ResourceAgent) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), " ")
	if len(parts) < 2 {
		return errors.New("expected at least type and name")
	}

	r.Type = parts[0]
	r.Name = parts[1]
	r.Attributes = map[string]string{}
	for _, arg := range parts[2:] {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			return errors.New("expected key=value pairs as arguments")
		}

		r.Attributes[kv[0]] = kv[1]
	}

	return nil
}

func (r ResourceAgent) MarshalText() (text []byte, err error) {
	args := make([]string, 0, len(r.Attributes))
	for k, v := range r.Attributes {
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}

	// Ensure consistent serialization order
	sort.Strings(args)

	return []byte(fmt.Sprintf("%s %s %s", r.Type, r.Name, strings.Join(args, " "))), nil
}

var _ toml.TextMarshaler = &ResourceAgent{}
var _ toml.TextUnmarshaler = &ResourceAgent{}
