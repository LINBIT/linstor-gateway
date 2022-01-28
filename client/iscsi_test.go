package client_test

import (
	"encoding/json"
	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func ipnet(str string) common.IpCidr {
	ip, err := common.ServiceIPFromString(str)
	if err != nil {
		panic(err)
	}
	return ip
}

func TestISCSIList(t *testing.T) {
	testcases := []struct {
		name     string
		response string
		actual   interface{}
		expected interface{}
	}{
		{
			response: `{"iqn":"iqn.2021-08.com.linbit:target1","resource_group":"DfltRscGrp","volumes":[{"number":1,"size_kib":47185920}],"service_ips":["10.43.6.223/16"],"status":{"state":"OK","service":"Started","primary":"test3","nodes":["test1","test2","test3"],"volumes":[{"number":1,"state":"OK"}]}}`,
			actual:   &iscsi.ResourceConfig{},
			expected: &iscsi.ResourceConfig{
				IQN:               iscsi.Iqn{"iqn.2021-08.com.linbit", "target1"},
				AllowedInitiators: nil,
				ResourceGroup:     "DfltRscGrp",
				Volumes: []common.VolumeConfig{
					{
						Number:  1,
						SizeKiB: 47185920,
					},
				},
				Username: "", Password: "",
				ServiceIPs: []common.IpCidr{ipnet("10.43.6.223/16")},
				Status: common.ResourceStatus{
					State:   common.ResourceStateOK,
					Service: common.ServiceStateStarted,
					Primary: "test3",
					Nodes:   []string{"test1", "test2", "test3"},
					Volumes: []common.VolumeState{
						{
							Number: 1,
							State:  common.ResourceStateOK,
						},
					},
				},
			},
		},
	}

	t.Parallel()
	for i := range testcases {
		tcase := &testcases[i]
		t.Run(reflect.TypeOf(tcase.expected).Name(), func(t *testing.T) {
			err := json.NewDecoder(strings.NewReader(tcase.response)).Decode(tcase.actual)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.Equal(t, tcase.expected, tcase.actual)
		})
	}
}
