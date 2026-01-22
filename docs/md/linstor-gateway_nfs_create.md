## linstor-gateway nfs create

Creates an NFS export

### Synopsis

Creates a highly available NFS export based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system under the
specified name and using the specified resource group.
After that it creates a drbd-reactor configuration to bring up a highly available NFS
export.

!!! NOTE that only one NFS resource can exist in a cluster.
To create multiple exports, create a single resource with multiple volumes.

```
linstor-gateway nfs create NAME SERVICE_IP [VOLUME_SIZE]... [flags]
```

### Examples

```
linstor-gateway nfs create example 192.168.211.122/24 2G
linstor-gateway nfs create restricted 10.10.22.44/16 2G --allowed-ips 10.10.0.0/16
linstor-gateway nfs create multi 172.16.16.55/24 1G 2G --export-path /music --export-path /movies

```

### Options

```
      --allowed-ips ip-cidr         Set the IP address mask of clients that are allowed access (default 0.0.0.0/0)
  -p, --export-path strings         Set the export path, relative to /srv/gateway-exports. Can be specified multiple times when creating more than one volume (default [/])
  -f, --filesystem string           File system type to use (ext4 or xfs) (default "ext4")
      --gross                       Make all size options specify gross size, i.e. the actual space used on disk
  -h, --help                        help for create
  -r, --resource-group string       LINSTOR resource group to use (default "DfltRscGrp")
      --resource-timeout duration   Timeout for waiting for the resource to become available (default 30s)
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

