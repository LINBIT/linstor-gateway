package reactor_test

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"

	"github.com/LINBIT/linstor-gateway/pkg/reactor"
)

func TestReactorConfig_UnmarshalText(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		cfg      string
		expected reactor.Config
		wantErr  bool
	}{{
		name: "empty",
		cfg:  `[[promoter]]`,
		expected: reactor.Config{
			Promoter: []reactor.PromoterConfig{{
				Resources: nil,
			}},
		},
	}, {
		name: "unexpected start type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = "start me"
`,
		wantErr: true,
	}, {
		name: "unexpected runner type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    runner = 1234
`,
		wantErr: true,
	}, {
		name: "unexpected on-drbd-demote-failure type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    on-drbd-demote-failure = true
`,
		wantErr: true,
	}, {
		name: "unexpected stop-services-on-exit type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    stop-services-on-exit = "not-a-boolean"
`,
		wantErr: true,
	}, {
		name: "unexpected target-as type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    target-as = 3.14
`,
		wantErr: true,
	}, {
		name: "unexpected start entry type",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = [ 1234 ]
`,
		wantErr: true,
	}, {
		name: "with systemd resources",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = [ "linstordb.mount" ]
`,
		expected: reactor.Config{
			Promoter: []reactor.PromoterConfig{{
				Resources: map[string]reactor.PromoterResourceConfig{
					"rsc1": {
						Start: []reactor.StartEntry{
							&reactor.SystemdService{
								Name: "linstordb.mount",
							},
						},
					},
				},
			}},
		},
	}, {
		name: "with resource agents",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = [ "ocf:heartbeat:IPaddr2 my_ip ip=1.2.3.4 cidr_netmask=24" ]
`,
		expected: reactor.Config{
			Promoter: []reactor.PromoterConfig{{
				Resources: map[string]reactor.PromoterResourceConfig{
					"rsc1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{
								Type: "ocf:heartbeat:IPaddr2",
								Name: "my_ip",
								Attributes: map[string]string{
									"ip":           "1.2.3.4",
									"cidr_netmask": "24",
								},
							},
						},
					},
				},
			}},
		},
	}, {
		name: "with mixed start entries",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = [ "ocf:heartbeat:IPaddr2 my_ip ip=1.2.3.4 cidr_netmask=24", "linstordb.mount" ]
`,
		expected: reactor.Config{
			Promoter: []reactor.PromoterConfig{{
				Resources: map[string]reactor.PromoterResourceConfig{
					"rsc1": {
						Start: []reactor.StartEntry{
							&reactor.ResourceAgent{
								Type: "ocf:heartbeat:IPaddr2",
								Name: "my_ip",
								Attributes: map[string]string{
									"ip":           "1.2.3.4",
									"cidr_netmask": "24",
								},
							},
							&reactor.SystemdService{
								Name: "linstordb.mount",
							},
						},
					},
				},
			}},
		},
	}, {
		name: "invalid ocf entry",
		cfg: `[[promoter]]
[promoter.resources]
  [promoter.resources.rsc1]
    start = [ "ocf:heartbeat:IPaddr2" ]
`,
		wantErr: true,
	}}

	for i := range testcases {
		tcase := &testcases[i]
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			cfg := reactor.Config{}
			err := toml.Unmarshal([]byte(tcase.cfg), &cfg)
			if tcase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tcase.expected, cfg)
			}
		})
	}
}

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
						Resources: nil,
					},
				},
			},
			expected: "\n[[promoter]]\n",
		},
		{
			name: "with-resources",
			cfg: reactor.Config{
				Promoter: []reactor.PromoterConfig{
					{
						Resources: map[string]reactor.PromoterResourceConfig{
							"rsc1": {Start: nil, Runner: "shell", OnDrbdDemoteFailure: "log", StopServicesOnExit: true, TargetAs: "BindsTo"},
							"rsc2": {},
						},
					},
				},
			},
			expected: `
[[promoter]]

  [promoter.resources]

    [promoter.resources.rsc1]
      on-drbd-demote-failure = "log"
      runner = "shell"
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
						Resources: map[string]reactor.PromoterResourceConfig{
							"rsc1": {
								Start: []reactor.StartEntry{
									&reactor.ResourceAgent{Type: "type1", Name: "name1", Attributes: map[string]string{"k1": "val1"}},
									&reactor.ResourceAgent{Type: "type2", Name: "name2", Attributes: map[string]string{}},
									&reactor.ResourceAgent{Type: "type3", Name: "name3", Attributes: map[string]string{"k3": "val3", "k3-2": "val3-2"}},
								},
							},
						},
					},
				},
			},
			expected: `
[[promoter]]

  [promoter.resources]

    [promoter.resources.rsc1]
      start = ["type1 name1 k1=val1", "type2 name2", "type3 name3 k3-2=val3-2 k3=val3"]
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
