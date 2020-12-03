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
  -c, --controller ip   Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help            help for start
  -i, --iqn string      Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
  -l, --lun int         Set the LUN Number (required) (default 1)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

