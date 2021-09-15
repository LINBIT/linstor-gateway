## linstor-gateway nfs create

Creates an NFS export

### Synopsis

Creates a highly available NFS export based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system under the
specified name and using the specified resource group.
After that it creates a drbd-reactor configuration to bring up a highly available NFS 
export.

```
linstor-gateway nfs create [flags]
```

### Examples

```
linstor-gateway nfs create --resource=example --service-ip=192.168.211.122/24 --allowed-ips=192.168.0.0/16 --resource-group=ssd_thin_2way --size=2G
```

### Options

```
      --allowed-ips ip-cidr     Set the IP address mask of clients that are allowed access (default ::1/64)
  -p, --export-path string      Set the export path (default "/")
  -h, --help                    help for create
  -r, --resource string         Set the resource name (required)
  -g, --resource-group string   Set the LINSTOR resource group name
      --service-ip ip-cidr      Set the service IP and netmask of the target (required) (default ::1/64)
      --size unit               Set a size (e.g, 1TiB) (default 1G)
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

