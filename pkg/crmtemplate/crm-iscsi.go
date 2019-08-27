package crmtemplate

const CRM_ISCSI = `<configuration>
    <resources>
      <primitive id="p_iscsi_{{.TargetName}}_ip" class="ocf" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_iscsi_{{.TargetName}}_ip-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_iscsi_{{.TargetName}}_ip-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="24" id="p_iscsi_{{.TargetName}}_ip-instance_attributes-cidr_netmask"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_{{.TargetName}}_ip-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}_ip-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}_ip-stop-0"/>
        </operations>
      </primitive>

      <primitive id="p_pblock_{{.TargetName}}" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_pblock_{{.TargetName}}-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_pblock_{{.TargetName}}-instance_attributes-ip"/>
	  <nvpair name="portno" value="3260" id="p_pblock_{{.TargetName}}-instance_attributes-portno"/>
	  <nvpair name="protocol" value="tcp" id="p_pblock_{{.TargetName}}-instance_attributes-protocol"/>
	  <nvpair name="action" value="block" id="p_pblock_{{.TargetName}}-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_pblock_{{.TargetName}}-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_pblock_{{.TargetName}}-stop-0"/>
        </operations>
        <meta_attributes id="p_pblock_{{.TargetName}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_pblock_{{.TargetName}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

      <primitive id="p_iscsi_{{.TargetName}}" class="ocf" provider="heartbeat" type="iSCSITarget">
        <instance_attributes id="p_iscsi_{{.TargetName}}-instance_attributes">
          <nvpair name="iqn" value="{{.Target.IQN}}" id="p_iscsi_{{.TargetName}}-instance_attributes-iqn"/>
          <nvpair name="incoming_username" value="{{$.Target.Username}}" id="p_iscsi_{{.TargetName}}-instance_attributes-incoming_username"/>
          <nvpair name="incoming_password" value="{{$.Target.Password}}" id="p_iscsi_{{.TargetName}}-instance_attributes-incoming_password"/>
          <nvpair name="portals" value="{{$.Target.Portals}}" id="p_iscsi_{{.TargetName}}-instance_attributes-portals"/>
          <nvpair name="tid" value="{{$.TID}}" id="p_iscsi_{{.TargetName}}-instance_attributes-tid"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}-stop-0"/>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_{{.TargetName}}-monitor-15"/>
        </operations>
        <meta_attributes id="p_iscsi_{{.TargetName}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_iscsi_{{.TargetName}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>

{{range .Target.LUNs}}
      <primitive id="p_iscsi_{{$.TargetName}}_lu{{.ID}}" class="ocf" provider="heartbeat" type="iSCSILogicalUnit">
        <instance_attributes id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-instance_attributes">
          <nvpair name="lun" value="{{.ID}}" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-instance_attributes-lun"/>
          <nvpair name="target_iqn" value="{{$.Target.IQN}}" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-instance_attributes-target_iqn"/>
          <nvpair name="path" value="{{$.Device}}" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-instance_attributes-path"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-stop-0"/>
          <op name="monitor" timeout="40" interval="15" id="p_iscsi_{{$.TargetName}}_lu{{.ID}}-monitor-15"/>
        </operations>
      </primitive>
{{end}}

      <primitive id="p_punblock_{{.TargetName}}" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_punblock_{{.TargetName}}-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_punblock_{{.TargetName}}-instance_attributes-ip"/>
	  <nvpair name="portno" value="3260" id="p_punblock_{{.TargetName}}-instance_attributes-portno"/>
	  <nvpair name="protocol" value="tcp" id="p_punblock_{{.TargetName}}-instance_attributes-protocol"/>
	  <nvpair name="action" value="unblock" id="p_punblock_{{.TargetName}}-instance_attributes-action"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_punblock_{{.TargetName}}-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_punblock_{{.TargetName}}-stop-0"/>
        </operations>
        <meta_attributes id="p_punblock_{{.TargetName}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_punblock_{{.TargetName}}-meta_attributes-target-role"/>
        </meta_attributes>
      </primitive>
    </resources>

    <constraints>
      <rsc_location id="lo_iscsi_{{.TargetName}}" resource-discovery="never">
        <resource_set id="lo_iscsi_{{.TargetName}}-0">
{{range $.Target.LUNs}}
          <resource_ref id="p_iscsi_{{$.TargetName}}_lu{{.ID}}"/>
{{end}}
          <resource_ref id="p_iscsi_{{.TargetName}}"/>
        </resource_set>
        <rule score="-INFINITY" id="lo_iscsi_{{.TargetName}}-rule">
{{range $i, $name := $.StorageNodes}}
		<expression attribute="#uname" operation="ne" value="{{$name}}" id="lo_iscsi_{{$.TargetName}}-rule-expression-{{$i}}"/>
{{end}}
        </rule>
      </rsc_location>
      <rsc_colocation id="co_pblock_{{.TargetName}}" score="INFINITY" rsc="p_pblock_{{.TargetName}}" with-rsc="p_iscsi_{{.TargetName}}_ip"/>
      <rsc_colocation id="co_iscsi_{{.TargetName}}" score="INFINITY" rsc="p_iscsi_{{.TargetName}}" with-rsc="p_pblock_{{.TargetName}}"/>
{{range $.Target.LUNs}}
      <rsc_colocation id="co_iscsi_{{$.TargetName}}_lu{{.ID}}" score="INFINITY" rsc="p_iscsi_{{$.TargetName}}_lu{{.ID}}" with-rsc="p_iscsi_{{$.TargetName}}"/>
{{end}}
      <rsc_colocation id="co_punblock_{{.TargetName}}" score="INFINITY" rsc="p_punblock_{{.TargetName}}" with-rsc="p_iscsi_{{.TargetName}}_ip"/>

{{range $_, $lu := $.Target.LUNs}}
      <rsc_location id="lo_iscsi_{{$.TargetName}}_lu{{$lu.ID}}" rsc="p_iscsi_{{$.TargetName}}_lu{{$lu.ID}}" resource-discovery="never">
        <rule score="0" id="lo_iscsi_{{$.TargetName}}_lu{{$lu.ID}}-rule">
	{{range $i, $name := $.StorageNodes}}
		<expression attribute="#uname" operation="ne" value="{{$name}}" id="lo_iscsi_{{$.TargetName}}_lu{{$lu.ID}}-rule-expression-{{$i}}"/>
	{{end}}
        </rule>
      </rsc_location>
{{end}}

      <rsc_order id="o_pblock_{{.TargetName}}" score="INFINITY" first="p_iscsi_{{.TargetName}}_ip" then="p_pblock_{{.TargetName}}"/>
      <rsc_order id="o_iscsi_{{.TargetName}}" score="INFINITY" first="p_pblock_{{.TargetName}}" then="p_iscsi_{{.TargetName}}"/>
{{range $.Target.LUNs}}
      <rsc_order id="o_iscsi_{{$.TargetName}}_lu{{.ID}}" score="INFINITY" first="p_iscsi_{{$.TargetName}}" then="p_iscsi_{{$.TargetName}}_lu{{.ID}}"/>
      <rsc_order id="o_punblock_{{$.TargetName}}" score="INFINITY" first="p_iscsi_{{$.TargetName}}_lu{{.ID}}" then="p_punblock_{{$.TargetName}}"/>
{{end}}
    </constraints>
</configuration>`
