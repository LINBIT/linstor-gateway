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
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

