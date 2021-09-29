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
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports
* [linstor-gateway iscsi create](linstor-gateway_iscsi_create.md)	 - Creates an iSCSI target
* [linstor-gateway iscsi delete](linstor-gateway_iscsi_delete.md)	 - Deletes an iSCSI target
* [linstor-gateway iscsi list](linstor-gateway_iscsi_list.md)	 - Lists iSCSI targets
* [linstor-gateway iscsi server](linstor-gateway_iscsi_server.md)	 - Starts a web server serving a REST API
* [linstor-gateway iscsi start](linstor-gateway_iscsi_start.md)	 - Starts an iSCSI target
* [linstor-gateway iscsi stop](linstor-gateway_iscsi_stop.md)	 - Stops an iSCSI target

