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
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

