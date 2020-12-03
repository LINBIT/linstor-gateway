## linstor-iscsi status

Displays the status of an iSCSI target or logical unit

### Synopsis

Triggers Pacemaker to probe the resoruce primitives of this iSCSI target.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
linstor-iscsi status --iqn=iqn.2019-08.com.linbit:example --lun=1

```
linstor-iscsi status [flags]
```

### Options

```
  -c, --controller ip   Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help            help for status
  -i, --iqn string      Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
  -l, --lun int         Set the iSCSI LU number (LUN) (default 1)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

