package linstorcontrol

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ResolveToken determines the LINSTOR API token from the given raw token and
// token file path. Setting both is an error. The token file may contain
// either the plain token or JSON of the form {"token": "..."}, matching the
// format of the auth.json file that the LINSTOR controller distributes to
// its satellites.
func ResolveToken(token, tokenFile string) (string, error) {
	if token != "" && tokenFile != "" {
		return "", fmt.Errorf("both a LINSTOR API token and a token file are configured; use only one")
	}
	if tokenFile == "" {
		return token, nil
	}
	contents, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	raw := strings.TrimSpace(string(contents))
	if !strings.HasPrefix(raw, "{") {
		return raw, nil
	}
	var parsed struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse token file %s: %w", tokenFile, err)
	}
	if parsed.Token == "" {
		return "", fmt.Errorf("token file %s does not contain a \"token\" key", tokenFile)
	}
	return parsed.Token, nil
}
