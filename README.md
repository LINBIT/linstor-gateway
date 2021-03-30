# linstor-gateway

`linstor-gateway` manages highly available iSCSI targets and NFS exports by leveraging on LINSTOR
and Pacemaker. Setting up LINSTOR - including storage pools and resource groups -
as well as Corosync and Pacemaker's properties are a prerequisite to use this tool.

# Building
Use a version of go that supports modules (>1.11). Then you can `go get` the code as usual.

```
go get github.com/LINBIT/linstor-gateway
```

# Requirements

## Pacemaker

A working Corosync/Pacemaker cluster is expected on the machine where linstor-gateway
is running.

The [drbd-attr](https://github.com/LINBIT/drbd-utils/blob/master/scripts/drbd-attr)
resource agent is required to run linstor-gateway. This is included in LINBIT's
drbd-utils package for Ubuntu based distributions, or the drbd-pacemaker package
on RHEL/CentOS.

linstor-gateway sets up all required Pacemaker resource and constraints by itself,
except for the LINSTOR controller resource.

## LINSTOR

A LINSTOR cluster is required to operate linstor-gateway. It is highly recommended
to run the LINSTOR controller as a Pacemaker resource. This needs to be configured
manually. Such a resource could look like the following:

```
primitive p_linstor-controller systemd:linstor-controller \
        op start interval=0 timeout=100s \
        op stop interval=0 timeout=100s \
        op monitor interval=30s timeout=100s
```

A storage pool needs to be created in LINSTOR. Also, a resource group for linstor-gateway
needs to be created.

## iSCSI

linstor-gateway uses Pacemaker's `ocf::heartbeat:iSCSITarget` resource agent for
its iSCSI integration, which requires an iSCSI implementation to be installed.
Using `targetcli` is recommended.

## NFS

There are a few requirements for NFS as well.

First, an NFS server needs to be started. The `nfsd` kernel module needs to be
loaded and the user-space NFS process needs to be running. The easiest way to
ensure that is to use

```
systemctl enable --now nfs-server
```

on all nodes.

The resource group that is used for the linstor-nfs create command needs to have
the `FileSystem/Type` attribute set. Configure this by doing
```
linstor resource-group set-property MyResourceGroup FileSystem/Type ext4
```

Note that currently only the `ext4` filesystem is supported.

# Documentation
Start by browsing the documentation for [linstor-iscsi](./docs/md/linstor-iscsi.md)
or [linstor-nfs](./docs/md/linstor-nfs.md).

The REST-API documentation can be found [here](https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway/).
