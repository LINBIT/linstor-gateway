## linstor-gateway nvme delete-volume

Delete a volume of an existing NVMe-oF target

### Synopsis

Delete a volume of an existing NVMe-oF target. The target needs to be stopped.

```
linstor-gateway nvme delete-volume NQN VOLUME_NR [flags]
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

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

