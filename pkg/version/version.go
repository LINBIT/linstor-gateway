package version

import "fmt"

// (potentially) set by makefile
var (
	Version   = "unknown"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func UserAgent() string {
	return fmt.Sprintf("linstor-gateway/%s-g%s", Version, GitCommit)
}
