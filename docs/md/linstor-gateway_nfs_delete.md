## linstor-gateway nfs delete

Deletes an NFS export

### Synopsis

Deletes an NFS export by stopping and deleting the drbd-reactor config
and removing the LINSTOR resources.

```
linstor-gateway nfs delete [flags]
```

### Examples

```
linstor-gateway nfs delete --resource=example
```

### Options

```
  -h, --help              help for delete
  -r, --resource string   Set the resource name (required)
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

