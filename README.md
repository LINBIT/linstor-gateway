# linstor-iscsi

`linstor-iscsi` manages highly available iSCSI targets by leveraging on LINSTOR
and Pacemaker. Setting up LINSTOR - including storage pools and resource groups -
as well as Corosync and Pacemaker's properties are a prerequisite to use this tool.

# Building
Use a version of go that supports modules (>1.11). Then you can `go get` the code as usual.

```
go get github.com/LINBIT/linstor-iscsi
```

# Requirements

## Pacemaker

A working Corosync/Pacemaker cluster is expected on the machine where linstor-iscsi
is running.

The [drbd-attr](https://github.com/LINBIT/drbd-utils/blob/master/scripts/drbd-attr)
resource agent is required to run linstor-iscsi. This is included in LINBIT's
drbd-utils package for Ubuntu based distributions, or the drbd-pacemaker package
on RHEL/CentOS.

linstor-iscsi sets up all required Pacemaker resource and constraints by itself,
except for the LINSTOR controller resource.

## LINSTOR

A LINSTOR cluster is required to operate linstor-iscsi. It is highly recommended
to run the LINSTOR controller as a Pacemaker resource. This needs to be configured
manually. Such a resource could look like the following:

```
primitive p_linstor-controller systemd:linstor-controller \
        op start interval=0 timeout=100s \
        op stop interval=0 timeout=100s \
        op monitor interval=30s timeout=100s
```

A storage pool needs to be created in LINSTOR. Also, a resource group for linstor-iscsi
needs to be created.

## iSCSI

linstor-iscsi uses Pacemaker's `ocf::heartbeat:iSCSITarget` resource agent, which
requires an iSCSI implementation to be installed. Using `targetcli` is recommended.

# Documentation
Start by browsing the documentation [here](./docs/md/linstor-iscsi.md).

The REST-API documentation can be found [here](https://app.swaggerhub.com/apis-docs/Linstor/linstor-iscsi/).
