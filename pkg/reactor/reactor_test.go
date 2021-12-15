package reactor

import (
	"github.com/LINBIT/golinstor/client"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestFilterConfigs(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		files           []client.ExternalFile
		expectedConfigs []PromoterConfig
		expectedPaths   []string
		wantErr         bool
	}{{
		name:            "empty files",
		files:           []client.ExternalFile{},
		expectedConfigs: []PromoterConfig{},
		expectedPaths:   []string{},
	}, {
		name: "one file",
		files: []client.ExternalFile{
			{
				Path: filepath.Join(promoterDir, "linstor-gateway-iscsi-target1.toml"),
				Content: []byte(`[[promoter]]
id = "iscsi-target1"
`),
			},
		},
		expectedConfigs: []PromoterConfig{{ID: "iscsi-target1", Resources: nil}},
		expectedPaths:   []string{filepath.Join(promoterDir, "linstor-gateway-iscsi-target1.toml")},
	}, {
		name: "one file with invalid contents",
		files: []client.ExternalFile{
			{
				Path:    filepath.Join(promoterDir, "linstor-gateway-iscsi-target1.toml"),
				Content: []byte(`don't know what this is, but it's not toml!`),
			},
		},
		wantErr: true,
	}, {
		name: "one relevant file",
		files: []client.ExternalFile{
			{
				Path: filepath.Join(promoterDir, "linstor-gateway-iscsi-target1.toml"),
				Content: []byte(`[[promoter]]
id = "iscsi-target1"
`),
			},
			{Path: "/some/other/file"},
			{Path: filepath.Join(promoterDir, "oops-not-the-right-pattern.toml")},
		},
		expectedConfigs: []PromoterConfig{{ID: "iscsi-target1", Resources: nil}},
		expectedPaths:   []string{filepath.Join(promoterDir, "linstor-gateway-iscsi-target1.toml")},
	}}

	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			configs, paths, err := filterConfigs(tcase.files)
			if tcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tcase.expectedConfigs, configs)
				assert.Equal(t, tcase.expectedPaths, paths)
			}
		})
	}
}
