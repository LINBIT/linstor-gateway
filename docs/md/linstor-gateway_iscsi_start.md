## linstor-gateway iscsi start

Starts an iSCSI target

### Synopsis

Sets the target role attribute of a Pacemaker primitive to started.
In case it does not start use your favourite pacemaker tools to analyze
the root cause.

```
linstor-gateway iscsi start [flags]
```

### Examples

```
linstor-gateway iscsi start --iqn=iqn.2019-08.com.linbit:example
```

### Options

```
  -h, --help      help for start
  -i, --iqn iqn   Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (default :)
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

