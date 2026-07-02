package linstorcontrol_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
)

func writeTokenFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
	return path
}

func TestResolveToken(t *testing.T) {
	t.Parallel()

	t.Run("no token configured", func(t *testing.T) {
		token, err := linstorcontrol.ResolveToken("", "")
		assert.NoError(t, err)
		assert.Equal(t, "", token)
	})

	t.Run("raw token", func(t *testing.T) {
		token, err := linstorcontrol.ResolveToken("secret123", "")
		assert.NoError(t, err)
		assert.Equal(t, "secret123", token)
	})

	t.Run("both token and token file", func(t *testing.T) {
		_, err := linstorcontrol.ResolveToken("secret123", "/some/path")
		assert.Error(t, err)
	})

	t.Run("plain text file", func(t *testing.T) {
		path := writeTokenFile(t, "secret123\n")
		token, err := linstorcontrol.ResolveToken("", path)
		assert.NoError(t, err)
		assert.Equal(t, "secret123", token)
	})

	t.Run("json file", func(t *testing.T) {
		path := writeTokenFile(t, `{"token": "secret123"}`)
		token, err := linstorcontrol.ResolveToken("", path)
		assert.NoError(t, err)
		assert.Equal(t, "secret123", token)
	})

	t.Run("json file without token key", func(t *testing.T) {
		path := writeTokenFile(t, `{"something": "else"}`)
		_, err := linstorcontrol.ResolveToken("", path)
		assert.Error(t, err)
	})

	t.Run("invalid json file", func(t *testing.T) {
		path := writeTokenFile(t, `{"token": `)
		_, err := linstorcontrol.ResolveToken("", path)
		assert.Error(t, err)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := linstorcontrol.ResolveToken("", filepath.Join(t.TempDir(), "nope"))
		assert.Error(t, err)
	})
}
