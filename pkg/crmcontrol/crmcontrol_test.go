package crmcontrol

import (
	"bytes"
	"net"
	"strings"
	"testing"

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

	if !cmp.Equal(config.Targets, expectedTargets) {
		t.Errorf("Targets are not equal")
		t.Errorf("Expected: %+v", expectedTargets)
		t.Errorf("Actual: %+v", config.Targets)
	}

	expectedLus := []*crmLu{
		&crmLu{
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

	expectedIPs := []*crmIP{
		&crmIP{
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

func TestModifyCrmTargetRole(t *testing.T) {
	expect := `<cib><configuration><resources>
			<primitive id="p_iscsi_example">
				<meta_attributes id="p_iscsi_example-meta_attributes">
					<nvpair name="target-role" value="Stopped" id="p_iscsi_example-meta_attributes-target-role"/>
				</meta_attributes>
			</primitive>
		</resources></configuration></cib>`

	cases := []struct {
		desc        string
		input       string
		expectError bool
	}{{
		desc: "nvpair present",
		input: `<cib><configuration><resources>
			<primitive id="p_iscsi_example">
				<meta_attributes id="p_iscsi_example-meta_attributes">
					<nvpair name="target-role" value="Started" id="p_iscsi_example-meta_attributes-target-role"/>
				</meta_attributes>
			</primitive>
		</resources></configuration></cib>`,
	}, {
		desc: "no nvpair present",
		input: `<cib><configuration><resources>
			<primitive id="p_iscsi_example">
				<meta_attributes id="p_iscsi_example-meta_attributes">
				</meta_attributes>
			</primitive>
		</resources></configuration></cib>`,
	}, {
		desc: "no meta_attributes present",
		input: `<cib><configuration><resources>
			<primitive id="p_iscsi_example">
			</primitive>
		</resources></configuration></cib>`,
	}, {
		desc: "no primitive present",
		input: `<cib><configuration><resources>
		</resources></configuration></cib>`,
		expectError: true,
	}}

	n := xmltest.Normalizer{OmitWhitespace: true}

	// store normalized version of expected XML
	var buf bytes.Buffer
	if err := n.Normalize(&buf, strings.NewReader(expect)); err != nil {
		t.Fatal(err)
	}
	normExpect := buf.String()

	for _, c := range cases {
		doc := xmltree.NewDocument()
		err := doc.ReadFromString(c.input)
		if err != nil {
			t.Fatal(err)
		}

		doc, err = modifyCrmTargetRole("p_iscsi_example", false, doc)
		if err != nil {
			if !c.expectError {
				t.Error("Unexpected error: ", err)
			}
			continue
		}

		if c.expectError {
			t.Error("Expected error")
			continue
		}

		actual, err := doc.WriteToString()
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if err := n.Normalize(&buf, strings.NewReader(actual)); err != nil {
			t.Fatal(err)
		}
		normActual := buf.String()

		if normActual != normExpect {
			t.Errorf("XML does not match (input '%s')", c.desc)
			t.Errorf("Expected: %s", normExpect)
			t.Errorf("Actual: %s", normActual)
		}
	}
}

func TestDissolveConstraints(t *testing.T) {
	xml := `<cib><configuration><constraints>
<rsc_location id="lo_iscsi_example" resource-discovery="never">
	<resource_set id="lo_iscsi_example-0">
		<resource_ref id="p_iscsi_example_lu1"/>
		<resource_ref id="p_iscsi_example"/>
	</resource_set>
	<rule score="-INFINITY" id="lo_iscsi_example-rule">
		<expression attribute="#uname" operation="ne" value="li0" id="lo_iscsi_example-rule-expression-0"/>
		<expression attribute="#uname" operation="ne" value="li1" id="lo_iscsi_example-rule-expression-1"/>
	</rule>
</rsc_location>
<rsc_colocation id="co_pblock_example" score="INFINITY" rsc="p_pblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_colocation id="co_iscsi_example" score="INFINITY" rsc="p_iscsi_example" with-rsc="p_pblock_example"/>
<rsc_colocation id="co_iscsi_example_lu1" score="INFINITY" rsc="p_iscsi_example_lu1" with-rsc="p_iscsi_example"/>
<rsc_colocation id="co_punblock_example" score="INFINITY" rsc="p_punblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_location id="lo_iscsi_example_lu1" rsc="p_iscsi_example_lu1" resource-discovery="never">
	<rule score="0" id="lo_iscsi_example_lu1-rule">
		<expression attribute="#uname" operation="ne" value="li0" id="lo_iscsi_example_lu1-rule-expression-0"/>
		<expression attribute="#uname" operation="ne" value="li1" id="lo_iscsi_example_lu1-rule-expression-1"/>
	</rule>
</rsc_location>
<rsc_order id="o_pblock_example" score="INFINITY" first="p_iscsi_example_ip" then="p_pblock_example"/>
<rsc_order id="o_iscsi_example" score="INFINITY" first="p_pblock_example" then="p_iscsi_example"/>
<rsc_order id="o_iscsi_example_lu1" score="INFINITY" first="p_iscsi_example" then="p_iscsi_example_lu1"/>
<rsc_order id="o_punblock_example" score="INFINITY" first="p_iscsi_example_lu1" then="p_punblock_example"/>
</constraints></configuration><status>
	<node_state><lrm id="171"><lrm_resources>
		<lrm_resource id="p_iscsi_example_ip" type="IPaddr2" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_pblock_example" type="portblock" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_iscsi_example" type="iSCSITarget" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_iscsi_example_lu1" type="iSCSILogicalUnit" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_punblock_example" type="portblock" class="ocf" provider="heartbeat"/>
	</lrm_resources></lrm></node_state>
</status></cib>`

	docRoot := xmltree.NewDocument()
	err := docRoot.ReadFromString(xml)
	if err != nil {
		t.Fatalf("Invalid XML in test data: %v", err)
	}

	cases := []struct {
		desc      string
		resources []string
		expect    string
	}{{
		desc:      "remove target",
		resources: []string{"p_iscsi_example"},
		expect: `<cib><configuration><constraints>
<rsc_colocation id="co_pblock_example" score="INFINITY" rsc="p_pblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_colocation id="co_punblock_example" score="INFINITY" rsc="p_punblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_location id="lo_iscsi_example_lu1" rsc="p_iscsi_example_lu1" resource-discovery="never">
	<rule score="0" id="lo_iscsi_example_lu1-rule">
		<expression attribute="#uname" operation="ne" value="li0" id="lo_iscsi_example_lu1-rule-expression-0"/>
		<expression attribute="#uname" operation="ne" value="li1" id="lo_iscsi_example_lu1-rule-expression-1"/>
	</rule>
</rsc_location>
<rsc_order id="o_pblock_example" score="INFINITY" first="p_iscsi_example_ip" then="p_pblock_example"/>
<rsc_order id="o_punblock_example" score="INFINITY" first="p_iscsi_example_lu1" then="p_punblock_example"/>
</constraints></configuration><status>
	<node_state><lrm id="171"><lrm_resources>
		<lrm_resource id="p_iscsi_example_ip" type="IPaddr2" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_pblock_example" type="portblock" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_iscsi_example_lu1" type="iSCSILogicalUnit" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_punblock_example" type="portblock" class="ocf" provider="heartbeat"/>
	</lrm_resources></lrm></node_state>
</status></cib>`,
	}, {
		desc:      "remove target, lu",
		resources: []string{"p_iscsi_example", "p_iscsi_example_lu1"},
		expect: `<cib><configuration><constraints>
<rsc_colocation id="co_pblock_example" score="INFINITY" rsc="p_pblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_colocation id="co_punblock_example" score="INFINITY" rsc="p_punblock_example" with-rsc="p_iscsi_example_ip"/>
<rsc_order id="o_pblock_example" score="INFINITY" first="p_iscsi_example_ip" then="p_pblock_example"/>
</constraints></configuration><status>
	<node_state><lrm id="171"><lrm_resources>
		<lrm_resource id="p_iscsi_example_ip" type="IPaddr2" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_pblock_example" type="portblock" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_punblock_example" type="portblock" class="ocf" provider="heartbeat"/>
	</lrm_resources></lrm></node_state>
</status></cib>`,
	}, {
		desc:      "remove target, lu, ip",
		resources: []string{"p_iscsi_example", "p_iscsi_example_lu1", "p_iscsi_example_ip"},
		expect: `<cib><configuration><constraints></constraints></configuration><status>
	<node_state><lrm id="171"><lrm_resources>
		<lrm_resource id="p_pblock_example" type="portblock" class="ocf" provider="heartbeat"/>
		<lrm_resource id="p_punblock_example" type="portblock" class="ocf" provider="heartbeat"/>
	</lrm_resources></lrm></node_state>
</status></cib>`,
	}, {
		desc:      "remove target, lu, ip, pblock",
		resources: []string{"p_iscsi_example", "p_iscsi_example_lu1", "p_iscsi_example_ip", "p_pblock_example", "p_punblock_example"},
		expect:    `<cib><configuration><constraints></constraints></configuration><status><node_state><lrm id="171"><lrm_resources></lrm_resources></lrm></node_state></status></cib>`,
	}}

	n := xmltest.Normalizer{OmitWhitespace: true}

	for _, c := range cases {
		// store normalized version of expected XML
		var buf bytes.Buffer
		if err := n.Normalize(&buf, strings.NewReader(c.expect)); err != nil {
			t.Fatal(err)
		}
		normExpect := buf.String()

		doc := docRoot.Copy()

		dissolveConstraints(doc, c.resources)

		actual, err := doc.WriteToString()
		if err != nil {
			t.Fatal(err)
		}

		buf.Reset()
		if err := n.Normalize(&buf, strings.NewReader(actual)); err != nil {
			t.Fatal(err)
		}
		normActual := buf.String()

		if normActual != normExpect {
			t.Errorf("XML does not match (input '%s')", c.desc)
			t.Errorf("Expected: %s", normExpect)
			t.Errorf("Actual: %s", normActual)
		}
	}
}

func TestGenerateCreateLuXML(t *testing.T) {
	expect := `<configuration>
    <resources>
      <primitive id="p_iscsi_example_ip" class="ocf" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_iscsi_example_ip-instance_attributes">
          <nvpair name="ip" value="192.168.1.1" id="p_iscsi_example_ip-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="24" id="p_iscsi_example_ip-instance_attributes-cidr_netmask"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_example_ip-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_iscsi_example_ip-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_example_ip-stop-0"/>
        </operations>
      </primitive>

      <primitive id="p_pblock_example" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_pblock_example-instance_attributes">
          <nvpair name="ip" value="192.168.1.1" id="p_pblock_example-instance_attributes-ip"/>
          <nvpair name="portno" value="3260" id="p_pblock_example-instance_attributes-portno"/>
          <nvpair name="protocol" value="tcp" id="p_pblock_example-instance_attributes-protocol"/>
          <nvpair name="action" value="block" id="p_pblock_example-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_pblock_example-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_pblock_example-stop-0"/>
        </operations>
        <meta_attributes id="p_pblock_example-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_pblock_example-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

      <primitive id="p_iscsi_example" class="ocf" provider="heartbeat" type="iSCSITarget">
        <instance_attributes id="p_iscsi_example-instance_attributes">
          <nvpair name="iqn" value="iqn.2019-08.com.linbit:example" id="p_iscsi_example-instance_attributes-iqn"/>
          <nvpair name="incoming_username" value="user" id="p_iscsi_example-instance_attributes-incoming_username"/>
          <nvpair name="incoming_password" value="password" id="p_iscsi_example-instance_attributes-incoming_password"/>
          <nvpair name="portals" value="192.168.1.1:3260" id="p_iscsi_example-instance_attributes-portals"/>
          <nvpair name="tid" value="0" id="p_iscsi_example-instance_attributes-tid"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_example-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_example-stop-0"/>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_example-monitor-15"/>
        </operations>
        <meta_attributes id="p_iscsi_example-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_iscsi_example-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

      <primitive id="p_iscsi_example_lu0" class="ocf" provider="heartbeat" type="iSCSILogicalUnit">
        <instance_attributes id="p_iscsi_example_lu0-instance_attributes">
          <nvpair name="lun" value="0" id="p_iscsi_example_lu0-instance_attributes-lun"/>
          <nvpair name="target_iqn" value="iqn.2019-08.com.linbit:example" id="p_iscsi_example_lu0-instance_attributes-target_iqn"/>
          <nvpair name="path" value="/dev/drbd1000" id="p_iscsi_example_lu0-instance_attributes-path"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_example_lu0-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_example_lu0-stop-0"/>
          <op name="monitor" timeout="40" interval="15" id="p_iscsi_example_lu0-monitor-15"/>
        </operations>
      </primitive>

      <primitive id="p_punblock_example" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_punblock_example-instance_attributes">
          <nvpair name="ip" value="192.168.1.1" id="p_punblock_example-instance_attributes-ip"/>
          <nvpair name="portno" value="3260" id="p_punblock_example-instance_attributes-portno"/>
          <nvpair name="protocol" value="tcp" id="p_punblock_example-instance_attributes-protocol"/>
          <nvpair name="action" value="unblock" id="p_punblock_example-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_punblock_example-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_punblock_example-stop-0"/>
        </operations>
        <meta_attributes id="p_punblock_example-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_punblock_example-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>
    </resources>

    <constraints>
      <rsc_location id="lo_iscsi_example" resource-discovery="never">
        <resource_set id="lo_iscsi_example-0">
          <resource_ref id="p_iscsi_example_lu0"/>
          <resource_ref id="p_iscsi_example"/>
        </resource_set>
        <rule score="-INFINITY" id="lo_iscsi_example-rule">
<expression attribute="#uname" operation="ne" value="node0" id="lo_iscsi_example-rule-expression-0"/><expression attribute="#uname" operation="ne" value="node1" id="lo_iscsi_example-rule-expression-1"/>
        </rule>
      </rsc_location>
      <rsc_colocation id="co_pblock_example" score="INFINITY" rsc="p_pblock_example" with-rsc="p_iscsi_example_ip"/>
      <rsc_colocation id="co_iscsi_example" score="INFINITY" rsc="p_iscsi_example" with-rsc="p_pblock_example"/>
      <rsc_colocation id="co_iscsi_example_lu0" score="INFINITY" rsc="p_iscsi_example_lu0" with-rsc="p_iscsi_example"/>
      <rsc_colocation id="co_punblock_example" score="INFINITY" rsc="p_punblock_example" with-rsc="p_iscsi_example_ip"/>

      <rsc_location id="lo_iscsi_example_lu0" rsc="p_iscsi_example_lu0" resource-discovery="never">
        <rule score="0" id="lo_iscsi_example_lu0-rule">
<expression attribute="#uname" operation="ne" value="node0" id="lo_iscsi_example_lu0-rule-expression-0"/><expression attribute="#uname" operation="ne" value="node1" id="lo_iscsi_example_lu0-rule-expression-1"/>
        </rule>
      </rsc_location>

      <rsc_order id="o_pblock_example" score="INFINITY" first="p_iscsi_example_ip" then="p_pblock_example"/>
      <rsc_order id="o_iscsi_example" score="INFINITY" first="p_pblock_example" then="p_iscsi_example"/>
      <rsc_order id="o_iscsi_example_lu0" score="INFINITY" first="p_iscsi_example" then="p_iscsi_example_lu0"/>
      <rsc_order id="o_punblock_example" score="INFINITY" first="p_iscsi_example_lu0" then="p_punblock_example"/>
    </constraints>
</configuration>`
	n := xmltest.Normalizer{OmitWhitespace: true}
	var buf bytes.Buffer
	if err := n.Normalize(&buf, strings.NewReader(expect)); err != nil {
		t.Fatal(err)
	}
	normExpect := buf.String()

	storageNodeList := []string{"node0", "node1"}
	iscsiTargetName := "example"
	ip := net.ParseIP("192.168.1.1")
	iscsiTargetIqn := "iqn.2019-08.com.linbit:example"
	lun := uint8(0)
	device := "/dev/drbd1000"
	username := "user"
	password := "password"
	portal := "192.168.1.1:3260"
	tid := int16(0)
	actual, err := generateCreateLuXML(storageNodeList, iscsiTargetName, ip,
		iscsiTargetIqn, lun, device, username, password, portal, tid)
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
