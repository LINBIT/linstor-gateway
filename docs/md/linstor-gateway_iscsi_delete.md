## linstor-gateway iscsi delete

Deletes an iSCSI target

### Synopsis

Deletes an iSCSI target by stopping and deleting the pacemaker resource
primitives and removing the linstor resources.

```
linstor-gateway iscsi delete [flags]
```

### Examples

```
linstor-gateway iscsi delete --iqn=iqn.2019-08.com.linbit:example --lun=1
```

### Options

```
  -h, --help      help for delete
  -i, --iqn iqn   The iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) of the target to delete. (default :)
  -l, --lun int   Set the LUN (default -1)
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

