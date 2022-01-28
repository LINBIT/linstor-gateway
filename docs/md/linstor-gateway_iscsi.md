## linstor-gateway iscsi

Manages Highly-Available iSCSI targets

### Synopsis

linstor-gateway iscsi manages highly available iSCSI targets by leveraging
LINSTOR and drbd-reactor. Setting up LINSTOR, including storage pools and resource groups,
as well as drbd-reactor is a prerequisite to use this tool.

### Options

```
  -h, --help   help for iscsi
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports
* [linstor-gateway iscsi add-volume](linstor-gateway_iscsi_add-volume.md)	 - Add a new logical unit to an existing iSCSI target
* [linstor-gateway iscsi create](linstor-gateway_iscsi_create.md)	 - Creates an iSCSI target
* [linstor-gateway iscsi delete](linstor-gateway_iscsi_delete.md)	 - Deletes an iSCSI target
* [linstor-gateway iscsi delete-volume](linstor-gateway_iscsi_delete-volume.md)	 - Delete a logical unit of an existing iSCSI target
* [linstor-gateway iscsi list](linstor-gateway_iscsi_list.md)	 - Lists iSCSI targets
* [linstor-gateway iscsi start](linstor-gateway_iscsi_start.md)	 - Starts an iSCSI target
* [linstor-gateway iscsi stop](linstor-gateway_iscsi_stop.md)	 - Stops an iSCSI target

