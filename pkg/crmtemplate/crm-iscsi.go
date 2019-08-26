package crmtemplate

const CRM_ISCSI = `<configuration>
    <resources>
      <primitive id="p_iscsi_{{.Target.Name}}_ip" class="ocf" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_iscsi_{{.Target.Name}}_ip-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_iscsi_{{.Target.Name}}_ip-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="24" id="p_iscsi_{{.Target.Name}}_ip-instance_attributes-cidr_netmask"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_{{.Target.Name}}_ip-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{.Target.Name}}_ip-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{.Target.Name}}_ip-stop-0"/>
        </operations>
      </primitive>

      <primitive id="p_pblock_{{.Target.Name}}" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_pblock_{{.Target.Name}}-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_pblock_{{.Target.Name}}-instance_attributes-ip"/>
	  <nvpair name="portno" value="3260" id="p_pblock_{{.Target.Name}}-instance_attributes-portno"/>
	  <nvpair name="protocol" value="tcp" id="p_pblock_{{.Target.Name}}-instance_attributes-protocol"/>
	  <nvpair name="action" value="block" id="p_pblock_{{.Target.Name}}-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_pblock_{{.Target.Name}}-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_pblock_{{.Target.Name}}-stop-0"/>
        </operations>
        <meta_attributes id="p_pblock_{{.Target.Name}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_pblock_{{.Target.Name}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

      <primitive id="p_iscsi_{{.Target.Name}}" class="ocf" provider="heartbeat" type="iSCSITarget">
        <instance_attributes id="p_iscsi_{{.Target.Name}}-instance_attributes">
          <nvpair name="iqn" value="{{.Target.IQN}}" id="p_iscsi_{{.Target.Name}}-instance_attributes-iqn"/>
          <nvpair name="incoming_username" value="{{$.Target.Username}}" id="p_iscsi_{{.Target.Name}}-instance_attributes-incoming_username"/>
          <nvpair name="incoming_password" value="{{$.Target.Password}}" id="p_iscsi_{{.Target.Name}}-instance_attributes-incoming_password"/>
          <nvpair name="portals" value="{{$.Target.Portals}}" id="p_iscsi_{{.Target.Name}}-instance_attributes-portals"/>
          <nvpair name="tid" value="{{$.TID}}" id="p_iscsi_{{.Target.Name}}-instance_attributes-tid"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{.Target.Name}}-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{.Target.Name}}-stop-0"/>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_{{.Target.Name}}-monitor-15"/>
        </operations>
        <meta_attributes id="p_iscsi_{{.Target.Name}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_iscsi_{{.Target.Name}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

{{range .Target.LUNs}}
      <primitive id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}" class="ocf" provider="heartbeat" type="iSCSILogicalUnit">
        <instance_attributes id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-instance_attributes">
          <nvpair name="lun" value="{{.ID}}" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-instance_attributes-lun"/>
          <nvpair name="target_iqn" value="{{$.Target.IQN}}" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-instance_attributes-target_iqn"/>
          <nvpair name="path" value="{{$.Device}}" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-instance_attributes-path"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-stop-0"/>
          <op name="monitor" timeout="40" interval="15" id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}-monitor-15"/>
        </operations>
      </primitive>
{{end}}

      <primitive id="p_punblock_{{.Target.Name}}" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_punblock_{{.Target.Name}}-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_punblock_{{.Target.Name}}-instance_attributes-ip"/>
	  <nvpair name="portno" value="3260" id="p_punblock_{{.Target.Name}}-instance_attributes-portno"/>
	  <nvpair name="protocol" value="tcp" id="p_punblock_{{.Target.Name}}-instance_attributes-protocol"/>
	  <nvpair name="action" value="unblock" id="p_punblock_{{.Target.Name}}-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_punblock_{{.Target.Name}}-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_punblock_{{.Target.Name}}-stop-0"/>
        </operations>
        <meta_attributes id="p_punblock_{{.Target.Name}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_punblock_{{.Target.Name}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>
    </resources>

    <constraints>
      <rsc_location id="lo_iscsi_{{.Target.Name}}" resource-discovery="never">
        <resource_set id="lo_iscsi_{{.Target.Name}}-0">
{{range $.Target.LUNs}}
          <resource_ref id="p_iscsi_{{$.Target.Name}}_lu{{.ID}}"/>
{{end}}
          <resource_ref id="p_iscsi_{{.Target.Name}}"/>
        </resource_set>
        <rule score="-INFINITY" id="lo_iscsi_{{.Target.Name}}-rule">
{{range $i, $name := $.StorageNodes}}
		<expression attribute="#uname" operation="ne" value="{{$name}}" id="lo_iscsi_{{$.Target.Name}}-rule-expression-{{$i}}"/>
{{end}}
        </rule>
      </rsc_location>
      <rsc_colocation id="co_pblock_{{.Target.Name}}" score="INFINITY" rsc="p_pblock_{{.Target.Name}}" with-rsc="p_iscsi_{{.Target.Name}}_ip"/>
      <rsc_colocation id="co_iscsi_{{.Target.Name}}" score="INFINITY" rsc="p_iscsi_{{.Target.Name}}" with-rsc="p_pblock_{{.Target.Name}}"/>
{{range $.Target.LUNs}}
      <rsc_colocation id="co_iscsi_{{$.Target.Name}}_lu{{.ID}}" score="INFINITY" rsc="p_iscsi_{{$.Target.Name}}_lu{{.ID}}" with-rsc="p_iscsi_{{$.Target.Name}}"/>
{{end}}
      <rsc_colocation id="co_punblock_{{.Target.Name}}" score="INFINITY" rsc="p_punblock_{{.Target.Name}}" with-rsc="p_iscsi_{{.Target.Name}}_ip"/>

{{range $_, $lu := $.Target.LUNs}}
      <rsc_location id="lo_iscsi_{{$.Target.Name}}_lu{{$lu.ID}}" rsc="p_iscsi_{{$.Target.Name}}_lu{{$lu.ID}}" resource-discovery="never">
        <rule score="0" id="lo_iscsi_{{$.Target.Name}}_lu{{$lu.ID}}-rule">
	{{range $i, $name := $.StorageNodes}}
		<expression attribute="#uname" operation="ne" value="{{$name}}" id="lo_iscsi_{{$.Target.Name}}_lu{{$lu.ID}}-rule-expression-{{$i}}"/>
	{{end}}
        </rule>
      </rsc_location>
{{end}}

      <rsc_order id="o_pblock_{{.Target.Name}}" score="INFINITY" first="p_iscsi_{{.Target.Name}}_ip" then="p_pblock_{{.Target.Name}}"/>
      <rsc_order id="o_iscsi_{{.Target.Name}}" score="INFINITY" first="p_pblock_{{.Target.Name}}" then="p_iscsi_{{.Target.Name}}"/>
{{range $.Target.LUNs}}
      <rsc_order id="o_iscsi_{{$.Target.Name}}_lu{{.ID}}" score="INFINITY" first="p_iscsi_{{$.Target.Name}}" then="p_iscsi_{{$.Target.Name}}_lu{{.ID}}"/>
      <rsc_order id="o_punblock_{{$.Target.Name}}" score="INFINITY" first="p_iscsi_{{$.Target.Name}}_lu{{.ID}}" then="p_punblock_{{$.Target.Name}}"/>
{{end}}
    </constraints>
</configuration>`
