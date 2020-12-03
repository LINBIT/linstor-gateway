## linstor-iscsi delete

Deletes an iSCSI target

### Synopsis

Deletes an iSCSI target by stopping and deleting the pacemaker resource
primitives and removing the linstor resources.

For example:
linstor-iscsi delete --iqn=iqn.2019-08.com.linbit:example --lun=1

```
linstor-iscsi delete [flags]
```

### Options

```
  -c, --controller ip   Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help            help for delete
  -i, --iqn string      Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
  -l, --lun int         Set the LUN Number (required) (default 1)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

