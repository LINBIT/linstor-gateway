package iscsi

import "testing"

func TestCheckIQN(t *testing.T) {
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
		descr:       "empty string",
		input:       "",
		expectError: true,
	}}

	for _, c := range cases {
		err := CheckIQN(c.input)
		if err != nil != c.expectError {
			t.Errorf("Test '%s': Unexpected error state: %v", c.descr, err)
			continue
		}
	}
}
