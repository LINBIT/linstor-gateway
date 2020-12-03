## linstor-nfs status

Displays the status of an NFS export

### Synopsis

Triggers Pacemaker to probe the resoruce primitives of this NFS export.
That means Pacemaker will run the status operation on the nodes where the
resource can run.
This makes sure that Pacemakers view of the world is updated to the state
of the world.

For example:
linstor-nfs status --resource=example

```
linstor-nfs status [flags]
```

### Options

```
  -c, --controller ip     Set the IP of the linstor controller (default 127.0.0.1)
  -h, --help              help for status
  -r, --resource string   Set the resource name (required)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-nfs](linstor-nfs.md)	 - Manages Highly-Available NFS exports

