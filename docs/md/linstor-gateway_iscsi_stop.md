## linstor-gateway iscsi stop

Stops an iSCSI target

### Synopsis

Sets the target role attribute of a Pacemaker primitive to stopped.
This causes pacemaker to stop the components of an iSCSI target.

For example:
linstor-gateway iscsi stop --iqn=iqn.2019-08.com.linbit:example

```
linstor-gateway iscsi stop [flags]
```

### Options

```
  -h, --help      help for stop
  -i, --iqn iqn   Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required) (default :)
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

