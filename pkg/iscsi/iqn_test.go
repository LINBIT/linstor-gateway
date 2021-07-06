package iscsi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
)

func TestCheckIQN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		descr       string
		input       string
		expectError bool
	}{{
		descr:       "default case",
		input:       "iqn.2019-08.com.linbit:example",
		expectError: false,
	}, {
		descr:       "missing date",
		input:       "iqn.com.linbit:example",
		expectError: true,
	}, {
		descr:       "missing domain",
		input:       "iqn.2019-08:example",
		expectError: true,
	}, {
		descr:       "missing unique part",
		input:       "iqn.2019-08.com.linbit",
		expectError: true,
	}, {
		descr:       "missing unique part, but with colon",
		input:       "iqn.2019-08.com.linbit:",
		expectError: true,
	}, {
		descr:       "invalid unique part, starts with digit",
		input:       "iqn.2019-08.com.linbit:123example",
		expectError: true,
	}, {
		descr:       "invalid unique part, too short",
		input:       "iqn.2019-08.com.linbit:e",
		expectError: true,
	}, {
		descr:       "contains _",
		input:       "iqn.2019-08.com.lin_bit:example",
		expectError: true,
	}, {
		descr:       "contains space",
		input:       "iqn.2019-08.com.linbit:exa mple",
		expectError: true,
	}, {
		descr:       "empty string",
		input:       "",
		expectError: true,
	}}

	for i := range cases {
		tcase := &cases[i]
		t.Run(tcase.descr, func(t *testing.T) {
			t.Parallel()

			iqn, err := iscsi.NewIqn(tcase.input)
			if !tcase.expectError {
				assert.NoError(t, err)
				assert.Equal(t, tcase.input, iqn.String())
			} else {
				assert.Error(t, err)
			}
		})
	}
}
