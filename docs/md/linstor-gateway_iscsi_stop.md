## linstor-gateway iscsi stop

Stops an iSCSI target

### Synopsis

Disables an iSCSI target, making it unavailable to initiators while not deleting it.

```
linstor-gateway iscsi stop IQN [flags]
```

### Examples

```
linstor-gateway iscsi stop iqn.2019-08.com.linbit:example
```

### Options

```
  -h, --help   help for stop
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

