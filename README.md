# LINSTOR Gateway

LINSTOR Gateway manages highly available iSCSI targets, NFS exports, and NVMe-oF
targets by leveraging [LINSTOR](https://github.com/LINBIT/linstor-server) and
[drbd-reactor](https://github.com/LINBIT/drbd-reactor). A working LINSTOR cluster
with drbd-reactor are a prerequisite to use this tool.

# Quick Start

1. Set up a [LINSTOR](https://github.com/LINBIT/linstor-server) cluster. Ensure
   you have a [storage pool](https://linbit.com/drbd-user-guide/linstor-guide-1_0-en/#s-storage_pools)
   as well as a [resource group](https://linbit.com/drbd-user-guide/linstor-guide-1_0-en/#s-linstor-resource-groups)
   for your data.
2. Set up [drbd-reactor](https://github.com/LINBIT/drbd-reactor). The daemon
   should be configured to reload automatically when the configuration changes – follow
   the steps from the [drbd-reactor README](https://github.com/LINBIT/drbd-reactor/#automatic-reload).
3. LINSTOR Gateway is packaged as a single binary. Download one of the
   [releases](https://github.com/LINBIT/linstor-gateway/releases), put it
   into `/usr/local/bin`, and you are ready to go.

# Requirements

## iSCSI

When an iSCSI target is created, LINSTOR Gateway requires an iSCSI
implementation to be installed on all nodes where the target may run.

Using `targetcli` is recommended.

```
dnf install targetcli
# or
apt install targetcli
```

## NFS

For NFS exports, an NFS server must be installed and started on all nodes
where the export may be located:

```
systemctl enable --now nfs-server
```

# Documentation
Start by browsing the documentation for the [linstor-gateway](./docs/md/linstor-gateway.md)
command line utility.

The REST API documentation can be found [here](https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway/).

# Building

If you want to test the latest unstable version of LINSTOR Gateway, you can build
the git version from sources:

```
git clone https://github.com/LINBIT/linstor-gateway
cd linstor-gateway
go build .
```
