package linstorcontrol

import (
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func resState(inUse bool) *client.ResourceState {
	return &client.ResourceState{
		InUse: &inUse,
	}
}

func volume(number int32, diskState string) client.Volume {
	return client.Volume{
		VolumeNumber: number,
		State: client.VolumeState{
			DiskState: diskState,
		},
	}
}

func TestStatusFromResources(t *testing.T) {
	defaultResourceDefinition := &client.ResourceDefinition{
		Name:              "test-resource",
		ResourceGroupName: "test-group",
		Props: map[string]string{
			"files/path/to/config.toml": "True",
		},
	}
	defaultResourceGroup := &client.ResourceGroup{
		Name:  "test-group",
		Props: nil,
		SelectFilter: client.AutoSelectFilter{
			PlaceCount: 2,
		},
	}
	type args struct {
		serviceCfgPath string
		definition     *client.ResourceDefinition
		group          *client.ResourceGroup
		resources      []client.ResourceWithVolumes
	}
	tests := []struct {
		name string
		args args
		want common.ResourceStatus
	}{
		{
			name: "two-replicas-one-diskless",
			args: args{
				serviceCfgPath: "/path/to/config.toml",
				resources: []client.ResourceWithVolumes{{
					Resource: client.Resource{Name: "test-resource", NodeName: "node1", State: resState(true)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}, {
					Resource: client.Resource{Name: "test-resource", NodeName: "node2", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}, {
					Resource: client.Resource{Name: "test-resource", NodeName: "node3", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "Diskless"),
						volume(1, "Diskless"),
					},
				}},
			},
			want: common.ResourceStatus{
				State:   common.ResourceStateOK,
				Service: common.ServiceStateStarted,
				Primary: "node1",
				Nodes:   []string{"node1", "node2", "node3"},
				Volumes: []common.VolumeState{
					{Number: 0, State: common.ResourceStateOK},
					{Number: 1, State: common.ResourceStateOK},
				},
			},
		}, {
			name: "degraded-one-replica-one-diskless",
			args: args{
				serviceCfgPath: "/path/to/config.toml",
				resources: []client.ResourceWithVolumes{{
					Resource: client.Resource{Name: "test-resource", NodeName: "node1", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}, {
					Resource: client.Resource{Name: "test-resource", NodeName: "node2", State: resState(true)},
					Volumes: []client.Volume{
						volume(0, "Diskless"),
						volume(1, "Diskless"),
					},
				}},
			},
			want: common.ResourceStatus{
				State:   common.ResourceStateDegraded,
				Service: common.ServiceStateStarted,
				Primary: "node2",
				Nodes:   []string{"node1", "node2"},
				Volumes: []common.VolumeState{
					{Number: 0, State: common.ResourceStateDegraded},
					{Number: 1, State: common.ResourceStateDegraded},
				},
			},
		}, {
			name: "unknown-no-resources",
			args: args{
				serviceCfgPath: "/path/to/config.toml",
				resources:      []client.ResourceWithVolumes{},
			},
			want: common.ResourceStatus{
				State:   common.Unknown,
				Service: common.ServiceStateStopped,
				Primary: "",
				Nodes:   []string{},
				Volumes: []common.VolumeState{},
			},
		}, {
			name: "config-not-deployed",
			args: args{
				serviceCfgPath: "/path/to/config.toml",
				resources: []client.ResourceWithVolumes{{
					Resource: client.Resource{Name: "test-resource", NodeName: "node1", State: resState(true)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}, {
					Resource: client.Resource{Name: "test-resource", NodeName: "node2", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}},
				definition: &client.ResourceDefinition{
					Name:              "test-resource",
					ResourceGroupName: "test-group",
					Props:             map[string]string{},
				},
			},
			want: common.ResourceStatus{
				State:   common.ResourceStateOK,
				Service: common.ServiceStateStopped,
				Primary: "node1",
				Nodes:   []string{"node1", "node2"},
				Volumes: []common.VolumeState{
					{Number: 0, State: common.ResourceStateOK},
					{Number: 1, State: common.ResourceStateOK},
				},
			},
		}, {
			name: "config-deployed-but-no-resource-in-use",
			args: args{
				serviceCfgPath: "/path/to/config.toml",
				resources: []client.ResourceWithVolumes{{
					Resource: client.Resource{Name: "test-resource", NodeName: "node1", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}, {
					Resource: client.Resource{Name: "test-resource", NodeName: "node2", State: resState(false)},
					Volumes: []client.Volume{
						volume(0, "UpToDate"),
						volume(1, "UpToDate"),
					},
				}},
			},
			want: common.ResourceStatus{
				State:   common.ResourceStateOK,
				Service: common.ServiceStateStopped,
				Primary: "",
				Nodes:   []string{"node1", "node2"},
				Volumes: []common.VolumeState{
					{Number: 0, State: common.ResourceStateOK},
					{Number: 1, State: common.ResourceStateOK},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.definition == nil {
				tt.args.definition = defaultResourceDefinition
			}
			if tt.args.group == nil {
				tt.args.group = defaultResourceGroup
			}
			got := StatusFromResources(tt.args.serviceCfgPath, tt.args.definition, tt.args.group, tt.args.resources)
			assert.Equal(t, tt.want, got)
		})
	}
}
