## linstor-gateway iscsi delete

Deletes an iSCSI target

### Synopsis

Deletes an iSCSI target by stopping and deleting the corresponding
drbd-reactor configuration and removing the LINSTOR resources. All logical units
of the target will be deleted.

```
linstor-gateway iscsi delete IQN... [flags]
```

### Examples

```
linstor-gateway iscsi delete iqn.2019-08.com.linbit:example
```

### Options

```
  -f, --force   Delete without prompting for confirmation
  -h, --help    help for delete
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

