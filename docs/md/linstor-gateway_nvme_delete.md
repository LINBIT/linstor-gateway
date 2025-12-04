## linstor-gateway nvme delete

Delete existing NVMe-oF targets

```
linstor-gateway nvme delete NQN... [flags]
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

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

