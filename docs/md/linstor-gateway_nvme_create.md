## linstor-gateway nvme create

Create a new NVMe-oF target

### Synopsis

Create a new NVMe-oF target

```
linstor-gateway nvme create NQN SERVICE_IP [VOLUME_SIZE]... [flags]
```

### Options

```
  -h, --help                    help for create
  -r, --resource-group string   resource group to use. (default "DfltRscGrp")
```

### Options inherited from parent commands

```
      --config string         Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
      --loglevel string       Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

