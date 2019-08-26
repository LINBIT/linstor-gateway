package crmcontrol

import (
	"testing"

	xmltree "github.com/beevik/etree"
)

func TestFindLinstorControllerName(t *testing.T) {
	cases := []struct {
		descr       string
		input       string
		expect      string
		expectError bool
	}{{
		descr: "good",
		input: `<cib><status><node_state><lrm>
			<lrm_resources>
				<lrm_resource type="linstor-controller" class="systemd">
					<lrm_rsc_op on_node="node1" operation="start"/>
				</lrm_resource>
			</lrm_resources>
		</lrm></node_state></status></cib>`,
		expect:      "node1",
		expectError: false,
	}, {
		descr: "no lrm_resource",
		input: `<cib><status><node_state><lrm>
			<lrm_resources>
			</lrm_resources>
		</lrm></node_state></status></cib>`,
		expect:      "",
		expectError: true,
	}}

	for _, c := range cases {
		doc := xmltree.NewDocument()
		err := doc.ReadFromString(c.input)
		if err != nil {
			t.Fatalf("Invalid XML in test data: %v", err)
		}

		name, err := findLinstorControllerName(doc)
		if c.expectError != (err != nil) {
			t.Errorf("Unexpected error state: %v", err)
		}

		if name != c.expect {
			t.Error("Name did not match")
			t.Errorf("Expected: %s", c.expect)
			t.Errorf("Actual: %s", name)
		}
	}
}
