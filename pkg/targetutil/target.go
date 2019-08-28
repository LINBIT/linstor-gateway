package targetutil

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
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
	regexIQN = `iqn\.\d{4}-[0-1][0-9]\..*\..*`

	// This format is mandated by LINSTOR. Since we use the unique part
	// directly for LINSTOR resource names, it needs to be compliant.
	regexResourceName = `[[:alpha:]][[:alnum:]]+`

	regexWWN = regexp.MustCompile(`^` + regexIQN + `:` + regexResourceName + `$`)
)

// Minimum valid LUN for a volume
// LUN 0 is reserved for the (i)SCSI controller
const MinVolumeLun = 1

// TargetConfig contains the information necessary for iSCSI targets.
type TargetConfig struct {
	IQN       string `json:"iqn,omitempty"`
	LUNs      []*LUN `json:"luns,omitempty"`
	ServiceIP net.IP `json:"service_ip,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Portals   string `json:"portals,omitempty"`
}

func NewTarget(cfg TargetConfig) (Target, error) {
	if err := CheckIQN(cfg.IQN); err != nil {
		return Target{}, err
	}

	return Target{cfg}, nil
}

func NewTargetMust(cfg TargetConfig) Target {
	t, err := NewTarget(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// Target wraps a TargetConfig
type Target struct {
	TargetConfig
}

type LUN struct {
	ID      uint8  `json:"id,omitempty"`
	SizeKiB uint64 `json:"size_kib,omitempty"`
}

func CheckIQN(iqn string) error {
	if strings.ContainsAny(iqn, "_ ") {
		return errors.New("IQN cannot contain the characters '_' (underscore) or ' ' (space)")
	}

	if !regexWWN.MatchString(iqn) {
		return fmt.Errorf("Given IQN ('%s') does not match the regular expression '%s'", iqn, regexWWN.String())
	}

	return nil
}

// ExtractTargetName extracts the target name from an IQN string.
// e.g., in "iqn.2019-07.org.demo.filserver:filestorage", the "filestorage" part.
func ExtractTargetName(iqn string) (string, error) {
	spl := strings.Split(iqn, ":")
	if len(spl) != 2 {
		return "", errors.New("Malformed argument '" + iqn + "'")
	}
	return spl[1], nil
}
