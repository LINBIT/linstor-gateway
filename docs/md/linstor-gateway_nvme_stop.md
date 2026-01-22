## linstor-gateway nvme stop

Stop a started NVMe-oF target

```
linstor-gateway nvme stop NQN... [flags]
```

### Options

```
  -h, --help                        help for stop
      --resource-timeout duration   Timeout for waiting for the resource to become unavailable (default 30s)
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

