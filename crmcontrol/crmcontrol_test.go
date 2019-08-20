package crmcontrol

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"

	xmltree "github.com/beevik/etree"
	log "github.com/sirupsen/logrus"
)

func TestParseConfiguration(t *testing.T) {
	xml := `<cib>
	  <configuration>
	    <resources>
	      <primitive id="p_iscsi_example" class="ocf" provider="heartbeat" type="iSCSITarget">
		<instance_attributes id="p_iscsi_example-instance_attributes">
		  <nvpair name="iqn" value="iqn.2019-08.com.libit:example" id="p_iscsi_example-instance_attributes-iqn"/>
		  <nvpair name="incoming_username" value="rck" id="p_iscsi_example-instance_attributes-incoming_username"/>
		  <nvpair name="incoming_password" value="rck" id="p_iscsi_example-instance_attributes-incoming_password"/>
		  <nvpair name="portals" value="192.168.122.181:3260" id="p_iscsi_example-instance_attributes-portals"/>
		  <nvpair name="tid" value="2" id="p_iscsi_example-instance_attributes-tid"/>
		</instance_attributes>
	      </primitive>
	      <primitive id="p_iscsi_example_lu1" class="ocf" provider="heartbeat" type="iSCSILogicalUnit">
		<instance_attributes id="p_iscsi_example_lu1-instance_attributes">
		  <nvpair name="lun" value="1" id="p_iscsi_example_lu1-instance_attributes-lun"/>
		  <nvpair name="target_iqn" value="iqn.2019-08.com.libit:example" id="p_iscsi_example_lu1-instance_attributes-target_iqn"/>
		  <nvpair name="path" value="/dev/drbd1001" id="p_iscsi_example_lu1-instance_attributes-path"/>
		</instance_attributes>
	      </primitive>
	      <primitive id="p_iscsi_example_ip" class="ocf" provider="heartbeat" type="IPaddr2">
		<instance_attributes id="p_iscsi_example_ip-instance_attributes">
		  <nvpair name="ip" value="192.168.122.181" id="p_iscsi_example_ip-instance_attributes-ip"/>
		  <nvpair name="cidr_netmask" value="24" id="p_iscsi_example_ip-instance_attributes-cidr_netmask"/>
		</instance_attributes>
	      </primitive>
	    </resources>
	  </configuration>
	</cib>`
	docRoot := xmltree.NewDocument()
	log.SetLevel(log.DebugLevel)
	err := docRoot.ReadFromString(xml)
	if err != nil {
		t.Errorf("Invalid XML in test data: %v", err)
	}

	config, err := ParseConfiguration(docRoot)
	if err != nil {
		t.Errorf("Error while parsing config: %v", err)
	}

	expectedTargets := []*crmTarget{
		&crmTarget{
			ID:       "p_iscsi_example",
			IQN:      "iqn.2019-08.com.libit:example",
			Username: "rck",
			Password: "rck",
			Portals:  "192.168.122.181:3260",
			Tid:      2,
		},
	}

	if !cmp.Equal(config.TargetList, expectedTargets) {
		t.Errorf("Targets are not equal")
		t.Errorf("Expected: %+v", expectedTargets)
		t.Errorf("Actual: %+v", config.TargetList)
	}

	expectedLus := []*crmLu{
		&crmLu{
			ID:     "p_iscsi_example_lu1",
			LUN:    1,
			Target: expectedTargets[0],
			Path:   "/dev/drbd1001",
		},
	}

	if !cmp.Equal(config.LuList, expectedLus) {
		t.Errorf("LUs are not equal")
		t.Errorf("Expected: %+v", expectedLus)
		t.Errorf("Actual: %+v", config.LuList)
	}

	expectedIPs := []*crmIP{
		&crmIP{
			ID:      "p_iscsi_example_ip",
			IP:      net.ParseIP("192.168.122.181"),
			Netmask: 24,
		},
	}

	if !cmp.Equal(config.IPList, expectedIPs) {
		t.Errorf("IPs are not equal")
		t.Errorf("Expected: %+v", expectedIPs)
		t.Errorf("Actual: %+v", config.IPList)
	}
}
