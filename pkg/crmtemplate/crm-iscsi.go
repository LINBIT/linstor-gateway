package crmtemplate

const CRM_ISCSI = `<configuration>
    <resources>
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
          <nvpair name="failure-timeout" value="60" id="p_pblock_{{.TargetName}}-meta_attributes-failure_timeout"/>
        </meta_attributes>
      </primitive>

      <primitive id="p_iscsi_{{.TargetName}}_ip" class="ocf" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_iscsi_{{.TargetName}}_ip-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_iscsi_{{.TargetName}}_ip-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="{{.Target.ServiceIPNetmask}}" id="p_iscsi_{{.TargetName}}_ip-instance_attributes-cidr_netmask"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_iscsi_{{.TargetName}}_ip-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}_ip-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{.TargetName}}_ip-stop-0"/>
        </operations>
        <meta_attributes id="p_iscsi_{{.TargetName}}_ip-meta_attributes">
          <nvpair name="failure-timeout" value="60" id="p_iscsi_{{.TargetName}}_ip-meta_attributes-failure_timeout"/>
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
          <nvpair name="failure-timeout" value="60" id="p_iscsi_{{.TargetName}}-meta_attributes-failure_timeout"/>
        </meta_attributes>
      </primitive>

{{range .Target.LUNs}}
{{$rsc := (printf "%s_lu%d" $.TargetName .ID)}}
      <primitive id="p_iscsi_{{$rsc}}" class="ocf" provider="heartbeat" type="iSCSILogicalUnit">
        <instance_attributes id="p_iscsi_{{$rsc}}-instance_attributes">
          <nvpair name="lun" value="{{.ID}}" id="p_iscsi_{{$rsc}}-instance_attributes-lun"/>
          <nvpair name="target_iqn" value="{{$.Target.IQN}}" id="p_iscsi_{{$rsc}}-instance_attributes-target_iqn"/>
          <nvpair name="path" value="{{$.Device}}" id="p_iscsi_{{$rsc}}-instance_attributes-path"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="40" interval="0" id="p_iscsi_{{$rsc}}-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_iscsi_{{$rsc}}-stop-0"/>
          <op name="monitor" timeout="40" interval="15" id="p_iscsi_{{$rsc}}-monitor-15"/>
        </operations>
        <meta_attributes id="p_iscsi_{{$rsc}}-meta_attributes">
          <nvpair name="failure-timeout" value="60" id="p_iscsi_{{$rsc}}-meta_attributes-failure_timeout"/>
        </meta_attributes>
      </primitive>
{{end}}

      <primitive id="p_punblock_{{.TargetName}}" class="ocf" provider="heartbeat" type="portblock">
        <instance_attributes id="p_punblock_{{.TargetName}}-instance_attributes">
          <nvpair name="ip" value="{{.Target.ServiceIP}}" id="p_punblock_{{.TargetName}}-instance_attributes-ip"/>
          <nvpair name="portno" value="3260" id="p_punblock_{{.TargetName}}-instance_attributes-portno"/>
          <nvpair name="protocol" value="tcp" id="p_punblock_{{.TargetName}}-instance_attributes-protocol"/>
          <nvpair name="action" value="unblock" id="p_punblock_{{.TargetName}}-instance_attributes-action"/>
          <nvpair name="tickle_sync_nodes" value="{{.StorageNodesList}}" id="p_punblock_{{.TargetName}}-instance_attributes-tickle_sync_nodes"/>
        </instance_attributes>
        <operations>
          <op name="start" timeout="20" interval="0" id="p_punblock_{{.TargetName}}-start-0"/>
          <op name="stop" timeout="20" interval="0" id="p_punblock_{{.TargetName}}-stop-0"/>
          <op name="monitor" timeout="20" interval="15" id="p_punblock_{{.TargetName}}-monitor-0"/>
        </operations>
        <meta_attributes id="p_punblock_{{.TargetName}}-meta_attributes">
          <nvpair name="target-role" value="Started" id="p_punblock_{{.TargetName}}-meta_attributes-target-role"/>
          <nvpair name="failure-timeout" value="60" id="p_punblock_{{.TargetName}}-meta_attributes-failure_timeout"/>
        </meta_attributes>
      </primitive>
      <clone id="drbd-attr-clone">
        <primitive id="drbd-attr" class="ocf" provider="linbit" type="drbd-attr"/>
      </clone>
    </resources>

    <constraints>
{{range $.Target.LUNs}}
{{$rsc := (printf "%s_lu%d" $.TargetName .ID)}}
      <rsc_location id="lo_iscsi_{{$rsc}}" rsc="p_iscsi_{{$rsc}}">
        <rule score="-INFINITY" id="lo_iscsi_{{$rsc}}-rule">
          <expression attribute="drbd-promotion-score-{{$rsc}}" operation="not_defined" id="lo_iscsi_{{$rsc}}-rule-expression"/>
        </rule>
        <rule score-attribute="drbd-promotion-score-{{$rsc}}" id="lo_iscsi_{{$rsc}}-rule-0">
          <expression attribute="drbd-promotion-score-{{$rsc}}" operation="defined" id="lo_iscsi_{{$rsc}}-rule-0-expression"/>
        </rule>
      </rsc_location>
{{end}}

      <rsc_colocation id="co_set_{{.TargetName}}" score="INFINITY">
        <resource_set id="co_set_{{.TargetName}}-0">
          <resource_ref id="p_pblock_{{.TargetName}}"/>
          <resource_ref id="p_iscsi_{{.TargetName}}_ip"/>
          <resource_ref id="p_iscsi_{{.TargetName}}"/>
        </resource_set>
        <resource_set sequential="false" id="co_set_{{.TargetName}}-1">
{{range $.Target.LUNs}}
{{$rsc := (printf "%s_lu%d" $.TargetName .ID)}}
          <resource_ref id="p_iscsi_{{$rsc}}"/>
{{end}}
        </resource_set>
        <resource_set id="co_set_{{.TargetName}}-2">
          <resource_ref id="p_punblock_{{.TargetName}}"/>
        </resource_set>
      </rsc_colocation>

      <rsc_order id="o_set_{{.TargetName}}" kind="Mandatory">
        <resource_set id="o_set_{{.TargetName}}-0">
          <resource_ref id="p_pblock_{{.TargetName}}"/>
          <resource_ref id="p_iscsi_{{.TargetName}}_ip"/>
          <resource_ref id="p_iscsi_{{.TargetName}}"/>
        </resource_set>
        <resource_set sequential="false" id="o_set_{{.TargetName}}-1">
{{range $.Target.LUNs}}
{{$rsc := (printf "%s_lu%d" $.TargetName .ID)}}
          <resource_ref id="p_iscsi_{{$rsc}}"/>
{{end}}
        </resource_set>
        <resource_set id="o_set_{{.TargetName}}-2">
          <resource_ref id="p_punblock_{{.TargetName}}"/>
        </resource_set>
      </rsc_order>
    </constraints>
</configuration>`
