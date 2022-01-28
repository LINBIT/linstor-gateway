## linstor-gateway iscsi delete-volume

Delete a logical unit of an existing iSCSI target

### Synopsis

Delete a logical unit of an existing iSCSI target. The target needs to be stopped.

```
linstor-gateway iscsi delete-volume IQN LU_NR [flags]
```

### Options

```
  -h, --help   help for delete-volume
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

