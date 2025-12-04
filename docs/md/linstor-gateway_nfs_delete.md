## linstor-gateway nfs delete

Deletes an NFS export

### Synopsis

Deletes an NFS export by stopping and deleting the drbd-reactor config
and removing the LINSTOR resources.

```
linstor-gateway nfs delete NAME [flags]
```

### Examples

```
linstor-gateway nfs delete example
```

### Options

```
  -f, --force   Delete without prompting for confirmation
  -h, --help    help for delete
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

