package upgrade

import "github.com/LINBIT/linstor-gateway/pkg/reactor"

// removeID unsets the ID field of the promoter config.
// Starting with drbd-reactor v1.2.0, configs no longer require an ID. It
// prints a deprecation warning when it is used, so remove the field.
func removeID(cfg *reactor.PromoterConfig) error {
	cfg.ID = ""
	return nil
}

func firstResourceId(cfg *reactor.PromoterConfig) string {
	for k := range cfg.Resources {
		return k
	}
	return ""
}
