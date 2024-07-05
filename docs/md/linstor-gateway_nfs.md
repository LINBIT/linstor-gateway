## linstor-gateway nfs

Manages Highly-Available NFS exports

### Synopsis

linstor-gateway nfs manages highly available NFS exports by leveraging LINSTOR
and drbd-reactor. A running LINSTOR cluster including storage pools and resource groups
is a prerequisite to use this tool.

NOTE that only one NFS resource can exist in a cluster.
See "help nfs create" for more information

### Options

```
  -h, --help   help for nfs
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports
* [linstor-gateway nfs create](linstor-gateway_nfs_create.md)	 - Creates an NFS export
* [linstor-gateway nfs delete](linstor-gateway_nfs_delete.md)	 - Deletes an NFS export
* [linstor-gateway nfs list](linstor-gateway_nfs_list.md)	 - Lists NFS resources
* [linstor-gateway nfs upgrade](linstor-gateway_nfs_upgrade.md)	 - Check existing resources and upgrade their configuration if necessary

