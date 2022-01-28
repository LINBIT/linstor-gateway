## linstor-gateway iscsi list

Lists iSCSI targets

### Synopsis

Lists the iSCSI targets created with this tool and provides an overview
about the existing drbd-reactor and linstor parts.

```
linstor-gateway iscsi list [flags]
```

### Examples

```
linstor-gateway iscsi list
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

