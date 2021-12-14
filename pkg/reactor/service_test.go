package reactor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemdService_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected SystemdService
		wantErr  bool
	}{{
		name: "empty string",
		text: "",
		expected: SystemdService{
			Name: "",
		},
	}, {
		name: "with name",
		text: "linstordb.mount",
		expected: SystemdService{
			Name: "linstordb.mount",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SystemdService{}
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

func TestSystemdService_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		service  SystemdService
		expected string
		wantErr  bool
	}{{
		name:     "empty",
		service:  SystemdService{},
		expected: "",
	}, {
		name: "with name",
		service: SystemdService{
			Name: "linstordb.mount",
		},
		expected: "linstordb.mount",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.service.MarshalText()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(text))
			}
		})
	}
}
