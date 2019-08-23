## linstor-iscsi stop

Stops an iSCSI target

### Synopsis

Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=0

```
linstor-iscsi stop [flags]
```

### Options

```
  -h, --help   help for stop
```

### Options inherited from parent commands

```
  -i, --iqn string        Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
      --loglevel string   Set the log level (as defined by logrus) (default "info")
  -l, --lun int           Set the LUN Number (required)
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

