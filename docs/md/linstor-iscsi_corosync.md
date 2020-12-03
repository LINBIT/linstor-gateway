## linstor-iscsi corosync

Generates a corosync config

### Synopsis

Generates a corosync config

For example:
linstor-iscsi corosync --ips="192.168.1.1,192.168.1.2"

```
linstor-iscsi corosync [flags]
```

### Options

```
      --cluster-name string   name of the cluster (default "mycluster")
  -h, --help                  help for corosync
      --ips ipSlice           comma seprated list of IPs (e.g., 1.2.3.4,1.2.3.5) (default [127.0.0.1])
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

