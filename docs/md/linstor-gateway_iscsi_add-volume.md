## linstor-gateway iscsi add-volume

Add a new logical unit to an existing iSCSI target

### Synopsis

Add a new logical unit to an existing iSCSI target. The target needs to be stopped.

```
linstor-gateway iscsi add-volume IQN LU_NR LU_SIZE [flags]
```

### Options

```
  -h, --help   help for add-volume
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

