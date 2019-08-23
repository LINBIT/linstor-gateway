## linstor-iscsi probe

Probes an iSCSI target

### Synopsis

Triggers Pacemaker to probe the resoruce primitives of this iSCSI target.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
linstor-iscsi probe --iqn=iqn.2019-08.com.linbit:example --lun=0

```
linstor-iscsi probe [flags]
```

### Options

```
  -h, --help   help for probe
```

### Options inherited from parent commands

```
  -i, --iqn string        Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
      --loglevel string   Set the log level (as defined by logrus) (default "info")
  -l, --lun int           Set the LUN Number (required)
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

