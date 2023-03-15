## linstor-gateway nvme create

Create a new NVMe-oF target

### Synopsis

Create a new NVMe-oF target. The NQN consists of <vendor>:nvme:<subsystem>.

```
linstor-gateway nvme create NQN SERVICE_IP VOLUME_SIZE [VOLUME_SIZE]... [flags]
```

### Examples

```
linstor-gateway nvme create linbit:nvme:example
```

### Options

```
      --gross                   Make all size options specify gross size, i.e. the actual space used on disk
  -h, --help                    help for create
  -r, --resource-group string   resource group to use. (default "DfltRscGrp")
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway nvme](linstor-gateway_nvme.md)	 - Manages Highly-Available NVME targets

