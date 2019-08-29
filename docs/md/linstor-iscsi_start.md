## linstor-iscsi start

Starts an iSCSI target

### Synopsis

Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.

For example:
linstor-iscsi start --iqn=iqn.2019-08.com.linbit:example --lun=1

```
linstor-iscsi start [flags]
```

### Options

```
  -h, --help   help for start
```

### Options inherited from parent commands

```
  -i, --iqn string        Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
      --loglevel string   Set the log level (as defined by logrus) (default "info")
  -l, --lun int           Set the LUN Number (required) (default 1)
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

