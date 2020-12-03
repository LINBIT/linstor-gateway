## linstor-nfs delete

Deletes an NFS export

### Synopsis

Deletes an NFS export by stopping and deleting the pacemaker resource
primitives and removing the linstor resources.

For example:
linstor-nfs delete --resource=example

```
linstor-nfs delete [flags]
```

### Options

```
  -c, --controller ip     Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help              help for delete
  -r, --resource string   Set the resource name (required)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-nfs](linstor-nfs.md)	 - Manages Highly-Available NFS exports

