## linstor-gateway nfs list

Lists NFS resources

### Synopsis

Lists the NFS resources created with this tool and provides an overview
about the existing LINSTOR resources and service status.

```
linstor-gateway nfs list [flags]
```

### Examples

```
linstor-gateway nfs list
```

### Options

```
  -c, --controller ip   Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help            help for list
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

