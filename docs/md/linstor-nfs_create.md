## linstor-nfs create

Creates an NFS export

### Synopsis

Creates a highly available NFS export based on LINSTOR and Pacemaker.
At first it creates a new resource within the linstor system under the
specified name and using the specified resource group.
After that it creates resource primitives in the Pacemaker cluster including
all necessary order and location constraints. The Pacemaker primites are
prefixed with p_, contain the resource name and a resource type postfix.

For example:
linstor-nfs create --resource=example --service-ip=192.168.211.122  \
 --allowed-ips=192.168.0.0/255.255.255.0 --resource-group=ssd_thin_2way --size=2G

Creates linstor resource example, volume 0 and
pacemaker primitives p_nfs_example_ip, p_nfs_example, p_nfs_example_export

```
linstor-nfs create [flags]
```

### Options

```
      --allowed-ips string      Set the IP address mask of clients that are allowed access
  -c, --controller ip           Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help                    help for create
  -r, --resource string         Set the resource name (required)
  -g, --resource-group string   Set the LINSTOR resource group name (default "default")
      --service-ip string       Set the service IP and netmask of the target (required) (default "127.0.0.1/8")
      --size unit               Set a size (e.g, 1TiB) (default 1G)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-nfs](linstor-nfs.md)	 - Manages Highly-Available NFS exports

