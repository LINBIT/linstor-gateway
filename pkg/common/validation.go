package common

import "fmt"

type ValidationError string

func (v ValidationError) Error() string {
	return fmt.Sprintf("invalid config: %s", string(v))
}
