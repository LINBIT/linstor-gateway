## linstor-gateway nvme start

Start a stopped NVMe-oF target

```
linstor-gateway nvme start NQN... [flags]
```

### Options

```
  -h, --help                        help for start
      --resource-timeout duration   Timeout for waiting for the resource to become available (default 30s)
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

