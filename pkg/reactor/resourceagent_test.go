package reactor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResourceAgent_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected ResourceAgent
		wantErr  bool
	}{{
		name:    "empty string",
		text:    "",
		wantErr: true,
	}, {
		name: "only type and name",
		text: "ocf:heartbeat:IPaddr2 my_ip",
		expected: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
		},
	}, {
		name: "with attributes",
		text: "ocf:heartbeat:IPaddr2 my_ip ip=1.2.3.4 netmask=24",
		expected: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"ip":      "1.2.3.4",
				"netmask": "24",
			},
		},
	}, {
		name:    "malformed attribute",
		text:    "ocf:heartbeat:IPaddr2 my_ip ip1.2.3.4",
		wantErr: true,
	}, {
		name: "attribute with space",
		text: "ocf:heartbeat:IPaddr2 my_ip title='my great IP'",
		expected: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"title": "my great IP",
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ResourceAgent{}
			err := r.UnmarshalText([]byte(tt.text))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, r)
			}
		})
	}
}

func TestResourceAgent_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		agent    ResourceAgent
		expected string
		wantErr  bool
	}{{
		name:    "empty",
		agent:   ResourceAgent{},
		wantErr: true,
	}, {
		name: "no type",
		agent: ResourceAgent{
			Name: "my_ip",
		},
		wantErr: true,
	}, {
		name: "no name",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
		},
		wantErr: true,
	}, {
		name: "only type and name",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
		},
		expected: "ocf:heartbeat:IPaddr2 my_ip",
	}, {
		name: "with attributes",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"ip":      "1.2.3.4",
				"netmask": "24",
			},
		},
		expected: "ocf:heartbeat:IPaddr2 my_ip ip=1.2.3.4 netmask=24",
	}, {
		name: "attributes are sorted",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"c": "3",
				"a": "1",
				"d": "4",
				"b": "2",
			},
		},
		expected: "ocf:heartbeat:IPaddr2 my_ip a=1 b=2 c=3 d=4",
	}, {
		name: "malformed attribute",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"ip": "1.2=3.4",
			},
		},
		expected: "ocf:heartbeat:IPaddr2 my_ip ip='1.2=3.4'",
	}, {
		name: "attribute with space",
		agent: ResourceAgent{
			Type: "ocf:heartbeat:IPaddr2",
			Name: "my_ip",
			Attributes: map[string]string{
				"title": "my great IP",
			},
		},
		expected: "ocf:heartbeat:IPaddr2 my_ip title='my great IP'",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.agent.MarshalText()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(text))
			}
		})
	}
}
