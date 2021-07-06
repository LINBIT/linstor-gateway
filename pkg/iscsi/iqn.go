package iscsi

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	// This format is currently dictated by the iSCSI target backend,
	// specifically the rtslib-fb library.
	// A notable difference in this implementation (which also differs from
	// RFC3720, where the IQN format is defined) is that we require the
	// "unique" part after the colon to be present.
	//
	// See also the source code of rtslib-fb for the original regex:
	// https://github.com/open-iscsi/rtslib-fb/blob/b5be390be961/rtslib/utils.py#L384
	regexIQN = `iqn\.\d{4}-[0-1][0-9]\.[^ _]*\.[^ _]*`

	// This format is mandated by LINSTOR. Since we use the unique part
	// directly for LINSTOR resource names, it needs to be compliant.
	regexResourceName = `[[:alpha:]][[:alnum:]]+`

	regexWWN = regexp.MustCompile(`^(` + regexIQN + `):(` + regexResourceName + `)$`)
)

type Iqn [2]string

func (i *Iqn) Set(s string) error {
	iqn, err := NewIqn(s)
	if err != nil {
		return err
	}

	*i = iqn
	return nil
}

func (i *Iqn) Type() string {
	return "iqn"
}

func (i Iqn) String() string {
	return fmt.Sprintf("%s:%s", i[0], i[1])
}

func (i *Iqn) WWN() string {
	return i[1]
}

func (i *Iqn) UnmarshalText(b []byte) error {
	match := regexWWN.FindStringSubmatch(string(b))

	if match == nil || len(match) != 3 {
		return invalidIqn(b)
	}

	*i = [2]string{match[1], match[2]}

	return nil
}

func (i *Iqn) UnmarshalJSON(b []byte) error {
	var raw string
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	return i.UnmarshalText([]byte(raw))
}

func (i *Iqn) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Iqn) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func NewIqn(s string) (Iqn, error) {
	var iqn Iqn
	err := iqn.UnmarshalText([]byte(s))
	if err != nil {
		return Iqn{}, err
	}

	return iqn, nil
}

type invalidIqn string

func (i invalidIqn) Error() string {
	return fmt.Sprintf("'%s' is not a valid IQN. expected format: iqn.YYYY-MM.DOTTED.DOMAIN.NAME:UNIQUE_RESOURCE_NAME", string(i))
}
