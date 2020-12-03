## linstor-iscsi stop

Stops an iSCSI target

### Synopsis

Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=1

```
linstor-iscsi stop [flags]
```

### Options

```
  -h, --help         help for stop
  -i, --iqn string   Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
  -l, --lun int      Set the LUN Number (required) (default 1)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

