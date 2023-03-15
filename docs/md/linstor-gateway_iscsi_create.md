## linstor-gateway iscsi create

Creates an iSCSI target

### Synopsis

Creates a highly available iSCSI target based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system, using the
specified resource group. The name of the linstor resources is derived
from the IQN's World Wide Name, which must be unique.
After that it creates a configuration for drbd-reactor to manage the
high availability primitives.

```
linstor-gateway iscsi create IQN SERVICE_IPS [VOLUME_SIZE]... [flags]
```

### Examples

```
linstor-gateway iscsi create iqn.2019-08.com.linbit:example 192.168.122.181/24 2G
```

### Options

```
      --allowed-initiators strings   Restrict which initiator IQNs are allowed to connect to the target
      --gross                        Make all size options specify gross size, i.e. the actual space used on disk
  -h, --help                         help for create
  -p, --password string              Set the password to use for CHAP authentication
  -g, --resource-group string        Set the LINSTOR resource group (default "DfltRscGrp")
  -u, --username string              Set the username to use for CHAP authentication
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

