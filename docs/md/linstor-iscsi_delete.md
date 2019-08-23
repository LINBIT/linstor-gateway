## linstor-iscsi delete

Deletes an iSCSI target

### Synopsis

Deletes an iSCSI target by stopping and deliting the pacemaker resource
primitives and removing the linstor resources.

For example:
linstor-iscsi delete --iqn=iqn.2019-08.com.linbit:example --lun=0

```
linstor-iscsi delete [flags]
```

### Options

```
  -c, --controller ip   Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help            help for delete
```

### Options inherited from parent commands

```
  -i, --iqn string        Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
      --loglevel string   Set the log level (as defined by logrus) (default "info")
  -l, --lun int           Set the LUN Number (required)
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

