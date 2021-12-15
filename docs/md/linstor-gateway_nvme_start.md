## linstor-gateway nvme start

Start a stopped NVMe-oF target

```
linstor-gateway nvme start NQN... [flags]
```

### Options

```
  -h, --help   help for start
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

