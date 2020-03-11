package crmtemplate

const CRM_NFS = `<configuration>
    <resources>
      <primitive id="p_nfs_{{.ResourceName}}_fs" class="ocf" provider="heartbeat" type="Filesystem">
        <instance_attributes id="p_nfs_{{.ResourceName}}_fs-instance_attributes">
          <nvpair name="device" value="{{.Device}}" id="p_nfs_{{.ResourceName}}_fs-instance_attributes-device"/>
          <nvpair name="directory" value="{{.Directory}}" id="p_nfs_{{.ResourceName}}_fs-instance_attributes-directory"/>
          <nvpair name="fstype" value="ext4" id="p_nfs_{{.ResourceName}}_fs-instance_attributes-fstype"/>
          <nvpair name="run_fsck" value="no" id="p_nfs_{{.ResourceName}}_fs-instance_attributes-run_fsck"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_nfs_{{.ResourceName}}_fs-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_fs-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_fs-stop-0"/>
        </operations>
      </primitive>

      <primitive id="p_nfs_{{.ResourceName}}_exp" class="ocf" provider="heartbeat" type="exportfs">
        <instance_attributes id="p_nfs_{{.ResourceName}}_exp-instance_attributes">
          <nvpair name="fsid" value="{{.FsId}}" id="p_nfs_{{.ResourceName}}_exp-instance_attributes-fsid"/>
          <nvpair name="directory" value="{{.Directory}}" id="p_nfs_{{.ResourceName}}_exp-instance_attributes-directory"/>
          <nvpair name="clientspec" value="{{.AllowedIPs}}" id="p_nfs_{{.ResourceName}}_exp-instance_attributes-clientspec"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_nfs_{{.ResourceName}}_exp-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_exp-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_exp-stop-0"/>
        </operations>
      </primitive>

      <primitive id="p_nfs_{{.ResourceName}}_ip" class="ocf" provider="heartbeat" type="IPaddr2">
        <instance_attributes id="p_nfs_{{.ResourceName}}_ip-instance_attributes">
          <nvpair name="ip" value="{{.ServiceIP}}" id="p_nfs_{{.ResourceName}}_ip-instance_attributes-ip"/>
          <nvpair name="cidr_netmask" value="{{.ServiceIPNetBits}}" id="p_nfs_{{.ResourceName}}_ip-instance_attributes-cidr_netmask"/>
        </instance_attributes>
        <operations>
          <op name="monitor" interval="15" timeout="40" id="p_nfs_{{.ResourceName}}_ip-monitor-15"/>
          <op name="start" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_ip-start-0"/>
          <op name="stop" timeout="40" interval="0" id="p_nfs_{{.ResourceName}}_ip-stop-0"/>
        </operations>
      </primitive>
    </resources>

    <constraints>
      <rsc_location id="lo_nfs_{{.ResourceName}}" resource-discovery="never">
        <resource_set id="lo_nfs_{{.ResourceName}}-0">
          <resource_ref id="p_nfs_{{.ResourceName}}_fs"/>
        </resource_set>
        <rule score="-INFINITY" id="lo_nfs_{{.ResourceName}}_fs-rule">
{{range $i, $name := $.StorageNodes}}
          <expression attribute="#uname" operation="ne" value="{{$name}}" id="lo_nfs_{{$.ResourceName}}_fs-rule-expression-{{$i}}"/>
{{end}}
        </rule>
      </rsc_location>

      <rsc_colocation id="co_nfs_{{.ResourceName}}" score="INFINITY">
        <resource_set id="co_nfs_{{.ResourceName}}-0">
          <resource_ref id="p_nfs_{{.ResourceName}}_fs"/>
          <resource_ref id="p_nfs_{{.ResourceName}}_exp"/>
          <resource_ref id="p_nfs_{{.ResourceName}}_ip"/>
        </resource_set>
      </rsc_colocation>

      <rsc_order id="o_nfs_{{.ResourceName}}">
        <resource_set id="o_nfs_{{.ResourceName}}-0" sequential="true">
          <resource_ref id="p_nfs_{{.ResourceName}}_fs"/>
          <resource_ref id="p_nfs_{{.ResourceName}}_exp"/>
          <resource_ref id="p_nfs_{{.ResourceName}}_ip"/>
        </resource_set>
      </rsc_order>
    </constraints>
</configuration>`
