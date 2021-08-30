## linstor-gateway nfs

Manages Highly-Available NFS exports

### Synopsis

linstor-gateway nfs manages higly available NFS exports by leveraging LINSTOR
and drbd-reactor. Setting linstor including storage pools and resource groups
as well as Corosync and Pacemaker's properties a prerequisite to use this tool.

### Options

```
  -h, --help              help for nfs
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports
* [linstor-gateway nfs create](linstor-gateway_nfs_create.md)	 - Creates an NFS export
* [linstor-gateway nfs delete](linstor-gateway_nfs_delete.md)	 - Deletes an NFS export
* [linstor-gateway nfs list](linstor-gateway_nfs_list.md)	 - Lists NFS resources
* [linstor-gateway nfs server](linstor-gateway_nfs_server.md)	 - Starts a web server serving a REST API

