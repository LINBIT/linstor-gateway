package reactor_test

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"

	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

func TestReactorConfig_MarshalText(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		cfg      reactor.Config
		expected string
	}{
		{
			name:     "empty",
			cfg:      reactor.Config{},
			expected: "",
		},
		{
			name: "empty-agents",
			cfg: reactor.Config{
				Promoter: []reactor.PromoterConfig{
					{
						ID:        "empty-agents",
						Resources: nil,
					},
				},
			},
			expected: "[[promoter]]\n  id = \"empty-agents\"\n",
		},
		{
			name: "with-resources",
			cfg: reactor.Config{
				Promoter: []reactor.PromoterConfig{
					{
						ID: "with-resources",
						Resources: map[string]reactor.PromoterResourceConfig{
							"rsc1": {Start: nil, Runner: "shell", OnDrbdDemoteFailure: "log", StopServicesOnExit: true, TargetAs: "BindsTo"},
							"rsc2": {},
						},
					},
				},
			},
			expected: `[[promoter]]
  id = "with-resources"
  [promoter.resources]
    [promoter.resources.rsc1]
      runner = "shell"
      on-drbd-demote-failure = "log"
      stop-services-on-exit = true
      target-as = "BindsTo"
    [promoter.resources.rsc2]
`,
		},
		{
			name: "resource-agents",
			cfg: reactor.Config{
				Promoter: []reactor.PromoterConfig{
					{
						ID: "resource-agents",
						Resources: map[string]reactor.PromoterResourceConfig{
							"rsc1": {
								Start: []reactor.ResourceAgent{
									{Type: "type1", Name: "name1", Attributes: map[string]string{"k1": "val1"}},
									{Type: "type2", Name: "name2", Attributes: map[string]string{}},
									{Type: "type3", Name: "name3", Attributes: map[string]string{"k3": "val3", "k3-2": "val3-2"}},
								},
							},
						},
					},
				},
			},
			expected: `[[promoter]]
  id = "resource-agents"
  [promoter.resources]
    [promoter.resources.rsc1]
      start = ["type1 name1 k1=val1", "type2 name2 ", "type3 name3 k3-2=val3-2 k3=val3"]
`,
		},
	}

	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			buffer := strings.Builder{}
			enc := toml.NewEncoder(&buffer)
			err := enc.Encode(&tcase.cfg)
			assert.NoError(t, err)
			assert.Equal(t, tcase.expected, buffer.String())
		})
	}
}
