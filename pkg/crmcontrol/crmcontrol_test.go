package crmcontrol

import (
	"bytes"
	"net"
	"sort"
	"strings"
	"testing"

	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	"github.com/google/go-cmp/cmp"
	"github.com/rsto/xmltest"

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
		t.Fatalf("Invalid XML in test data: %v", err)
	}

	config, err := ParseConfiguration(docRoot)
	if err != nil {
		t.Errorf("Error while parsing config: %v", err)
		return
	}

	expectedTargets := []*Target{
		&Target{
			ID:       "p_iscsi_example",
			IQN:      "iqn.2019-08.com.libit:example",
			Username: "rck",
			Password: "rck",
			Portals:  "192.168.122.181:3260",
			Tid:      2,
		},
	}

	if !cmp.Equal(config.Targets, expectedTargets) {
		t.Errorf("Targets are not equal")
		t.Errorf("Expected: %+v", expectedTargets)
		t.Errorf("Actual: %+v", config.Targets)
	}

	expectedLus := []*Lu{
		&Lu{
			ID:     "p_iscsi_example_lu1",
			LUN:    1,
			Target: expectedTargets[0],
			Path:   "/dev/drbd1001",
		},
	}

	if !cmp.Equal(config.LUs, expectedLus) {
		t.Errorf("LUs are not equal")
		t.Errorf("Expected: %+v", expectedLus)
		t.Errorf("Actual: %+v", config.LUs)
	}

	expectedIPs := []*IP{
		&IP{
			ID:      "p_iscsi_example_ip",
			IP:      net.ParseIP("192.168.122.181"),
			Netmask: 24,
		},
	}

	if !cmp.Equal(config.IPs, expectedIPs) {
		t.Errorf("IPs are not equal")
		t.Errorf("Expected: %+v", expectedIPs)
		t.Errorf("Actual: %+v", config.IPs)
	}
}

func TestGenerateCrmObjectNames(t *testing.T) {
	log.SetLevel(log.WarnLevel)
	expect := []string{"p_iscsi_example_ip",
		"p_pblock_example",
		"p_iscsi_example",
		"p_iscsi_example_lu1",
		"p_iscsi_example_lu105",
		"p_iscsi_example_lu12",
		"p_punblock_example",
	}
	actual := generateCrmObjectNames("example", []uint8{1, 105, 12})

	if !cmp.Equal(expect, actual) {
		t.Errorf("Generated object names are wrong")
		t.Errorf("Expected: %s", expect)
		t.Errorf("Actual: %s", actual)
	}
}

func TestGenerateCreateLuXML(t *testing.T) {
	expect := `<configuration>
    <resources>
      <primitive class="ocf" id="p_iscsi_example_ip" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_iscsi_example_ip-instance_attributes">
          <nvpair id="p_iscsi_example_ip-instance_attributes-ip" name="ip" value="192.168.1.1" />
          <nvpair id="p_iscsi_example_ip-instance_attributes-cidr_netmask" name="cidr_netmask" value="16" />
        </instance_attributes>
        <operations>
          <op id="p_iscsi_example_ip-monitor-15" interval="15" name="monitor" timeout="40" />
          <op id="p_iscsi_example_ip-start-0" interval="0" name="start" timeout="40" />
          <op id="p_iscsi_example_ip-stop-0" interval="0" name="stop" timeout="40" />
        </operations>
      </primitive>
      <primitive class="ocf" id="p_pblock_example" provider="heartbeat" type="portblock">
        <instance_attributes id="p_pblock_example-instance_attributes">
          <nvpair id="p_pblock_example-instance_attributes-ip" name="ip" value="192.168.1.1" />
          <nvpair id="p_pblock_example-instance_attributes-portno" name="portno" value="3260" />
          <nvpair id="p_pblock_example-instance_attributes-protocol" name="protocol" value="tcp" />
          <nvpair id="p_pblock_example-instance_attributes-action" name="action" value="block" />
        </instance_attributes>
        <operations>
          <op id="p_pblock_example-start-0" interval="0" name="start" timeout="20" />
          <op id="p_pblock_example-stop-0" interval="0" name="stop" timeout="20" />
        </operations>
        <meta_attributes id="p_pblock_example-meta_attributes">
          <nvpair id="p_pblock_example-meta_attributes-target-role" name="target-role" value="Started" />
        </meta_attributes>
      </primitive>
      <primitive class="ocf" id="p_iscsi_example" provider="heartbeat" type="iSCSITarget">
        <instance_attributes id="p_iscsi_example-instance_attributes">
          <nvpair id="p_iscsi_example-instance_attributes-iqn" name="iqn" value="iqn.2019-08.com.linbit:example" />
          <nvpair id="p_iscsi_example-instance_attributes-incoming_username" name="incoming_username" value="user" />
          <nvpair id="p_iscsi_example-instance_attributes-incoming_password" name="incoming_password" value="password" />
          <nvpair id="p_iscsi_example-instance_attributes-portals" name="portals" value="192.168.1.1:3260" />
          <nvpair id="p_iscsi_example-instance_attributes-tid" name="tid" value="0" />
        </instance_attributes>
        <operations>
          <op id="p_iscsi_example-start-0" interval="0" name="start" timeout="40" />
          <op id="p_iscsi_example-stop-0" interval="0" name="stop" timeout="40" />
          <op id="p_iscsi_example-monitor-15" interval="15" name="monitor" timeout="40" />
        </operations>
        <meta_attributes id="p_iscsi_example-meta_attributes">
          <nvpair id="p_iscsi_example-meta_attributes-target-role" name="target-role" value="Started" />
        </meta_attributes>
      </primitive>
      <primitive class="ocf" id="p_iscsi_example_lu0" provider="heartbeat" type="iSCSILogicalUnit">
        <instance_attributes id="p_iscsi_example_lu0-instance_attributes">
          <nvpair id="p_iscsi_example_lu0-instance_attributes-lun" name="lun" value="0" />
          <nvpair id="p_iscsi_example_lu0-instance_attributes-target_iqn" name="target_iqn" value="iqn.2019-08.com.linbit:example" />
          <nvpair id="p_iscsi_example_lu0-instance_attributes-path" name="path" value="/dev/drbd1000" />
        </instance_attributes>
        <operations>
          <op id="p_iscsi_example_lu0-start-0" interval="0" name="start" timeout="40" />
          <op id="p_iscsi_example_lu0-stop-0" interval="0" name="stop" timeout="40" />
          <op id="p_iscsi_example_lu0-monitor-15" interval="15" name="monitor" timeout="40" />
        </operations>
      </primitive>
      <primitive class="ocf" id="p_punblock_example" provider="heartbeat" type="portblock">
        <instance_attributes id="p_punblock_example-instance_attributes">
          <nvpair id="p_punblock_example-instance_attributes-ip" name="ip" value="192.168.1.1" />
          <nvpair id="p_punblock_example-instance_attributes-portno" name="portno" value="3260" />
          <nvpair id="p_punblock_example-instance_attributes-protocol" name="protocol" value="tcp" />
          <nvpair id="p_punblock_example-instance_attributes-action" name="action" value="unblock" />
          <nvpair id="p_punblock_example-instance_attributes-tickle_sync_nodes" name="tickle_sync_nodes" value="node0,node1" />
        </instance_attributes>
        <operations>
          <op id="p_punblock_example-start-0" interval="0" name="start" timeout="20" />
          <op id="p_punblock_example-stop-0" interval="0" name="stop" timeout="20" />
          <op id="p_punblock_example-monitor-0" name="monitor" timeout="20" interval="15" />
        </operations>
        <meta_attributes id="p_punblock_example-meta_attributes">
          <nvpair id="p_punblock_example-meta_attributes-target-role" name="target-role" value="Started" />
        </meta_attributes>
      </primitive>
      <clone id="drbd-attr-clone">
        <primitive class="ocf" id="drbd-attr" provider="linbit" type="drbd-attr" />
      </clone>
    </resources>
    <constraints>
      <rsc_location id="lo_iscsi_example_lu0" rsc="p_iscsi_example_lu0">
        <rule id="lo_iscsi_example_lu0-rule" score="-INFINITY">
          <expression attribute="drbd-promotion-score-example_lu0" id="lo_iscsi_example_lu0-rule-expression" operation="not_defined" />
        </rule>
        <rule id="lo_iscsi_example_lu0-rule-0" score-attribute="drbd-promotion-score-example_lu0">
          <expression attribute="drbd-promotion-score-example_lu0" id="lo_iscsi_example_lu0-rule-0-expression" operation="defined" />
        </rule>
      </rsc_location>
      <rsc_colocation id="co_pblock_example" rsc="p_pblock_example" score="INFINITY" with-rsc="p_iscsi_example_ip" />
      <rsc_colocation id="co_iscsi_example" rsc="p_iscsi_example" score="INFINITY" with-rsc="p_pblock_example" />
      <rsc_colocation id="co_iscsi_example_lu0" rsc="p_iscsi_example_lu0" score="INFINITY" with-rsc="p_iscsi_example" />
      <rsc_colocation id="co_punblock_example" rsc="p_punblock_example" score="INFINITY" with-rsc="p_iscsi_example_ip" />
      <rsc_order first="p_iscsi_example_ip" id="o_pblock_example" score="INFINITY" then="p_pblock_example" />
      <rsc_order first="p_pblock_example" id="o_iscsi_example" score="INFINITY" then="p_iscsi_example" />
      <rsc_order first="p_iscsi_example" id="o_iscsi_example_lu0" score="INFINITY" then="p_iscsi_example_lu0" />
      <rsc_order first="p_iscsi_example_lu0" id="o_punblock_example" score="INFINITY" then="p_punblock_example" />
    </constraints>
  </configuration>
	`
	n := xmltest.Normalizer{OmitWhitespace: true}
	var buf bytes.Buffer
	if err := n.Normalize(&buf, strings.NewReader(expect)); err != nil {
		t.Fatal(err)
	}
	normExpect := buf.String()

	storageNodes := []string{"node0", "node1"}
	device := "/dev/drbd1000"
	tid := int16(0)

	target := targetutil.NewTargetMust(targetutil.TargetConfig{
		IQN:              "iqn.2019-08.com.linbit:example",
		LUNs:             []*targetutil.LUN{&targetutil.LUN{ID: 0}},
		ServiceIP:        net.ParseIP("192.168.1.1"),
		ServiceIPNetmask: 16,
		Username:         "user",
		Password:         "password",
		Portals:          "192.168.1.1:3260",
	})

	actual, err := generateCreateLuXML(target, storageNodes, device, tid)
	if err != nil {
		t.Error(err)
		return
	}

	buf.Reset()
	if err := n.Normalize(&buf, strings.NewReader(actual)); err != nil {
		t.Fatal(err)
	}
	normActual := buf.String()

	if normActual != normExpect {
		t.Error("XML does not match")
		t.Errorf("Expected: %s", normExpect)
		t.Errorf("Actual: %s", normActual)
	}
}

func TestGetIDsToDelete(t *testing.T) {
	xml := `<cib><configuration><resources>
	<primitive id="p_iscsi_example" type="iSCSITarget"/>
	<primitive id="p_iscsi_example_lu1" type="iSCSILogicalUnit"/>
	<primitive id="p_iscsi_example_lu2" type="iSCSILogicalUnit"/>
	<primitive id="p_iscsi_example_ip" type="IPaddr2"/>
	<primitive id="p_pblock_example" type="portblock"/>
	<primitive id="p_punblock_example" type="portblock"/>

	<primitive id="p_iscsi_example2" type="iSCSITarget"/>
	<primitive id="p_iscsi_example2_lu1" type="iSCSILogicalUnit"/>
	<primitive id="p_iscsi_example2_ip" type="IPaddr2"/>
	<primitive id="p_pblock_example2" type="portblock"/>
	<primitive id="p_punblock_example2" type="portblock"/>
</resources></configuration></cib>`

	var cib cib.CIB
	cib.Doc = xmltree.NewDocument()
	err := cib.Doc.ReadFromString(xml)
	if err != nil {
		t.Fatalf("Invalid XML in test data: %v", err)
	}

	cases := []struct {
		name   string
		lun    uint8
		expect []string
	}{{
		name: "example",
		lun:  1,
		// we only expect the LU because a second LU is present
		expect: []string{"p_iscsi_example_lu1"},
	}, {
		name: "example2",
		lun:  1,
		// we expect everything to be deleted because LU1 is the last LU
		expect: []string{"p_iscsi_example2", "p_iscsi_example2_lu1",
			"p_iscsi_example2_ip", "p_pblock_example2",
			"p_punblock_example2"},
	}}

	for _, c := range cases {
		ids, err := getIDsToDelete(&cib, c.name, c.lun)
		if err != nil {
			t.Error(err)
			return
		}

		sort.Strings(ids)
		sort.Strings(c.expect)

		if !cmp.Equal(ids, c.expect) {
			t.Errorf("IDs do not match for input %s", c.name)
			t.Errorf("Expected: %v", c.expect)
			t.Errorf("Actual:   %v", ids)
		}
	}
}
