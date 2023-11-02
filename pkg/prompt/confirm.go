package prompt

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Confirm displays a prompt `s` to the user and returns true if the user confirmed,
// false if not.
// If the lower cased, trimmed input is equal to 'y', that is considered to be
// a confirmation. Any other input value will return false.
func Confirm(s string) bool {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/N]: ", s)

	res, err := r.ReadString('\n')
	if err != nil {
		logrus.Error(err)
		return false
	}

	return strings.ToLower(strings.TrimSpace(res)) == "y"
}
