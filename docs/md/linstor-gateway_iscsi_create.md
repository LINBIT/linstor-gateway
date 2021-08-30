## linstor-gateway iscsi create

Creates an iSCSI target

### Synopsis

Creates a highly available iSCSI target based on LINSTOR and drbd-reactor.
At first it creates a new resource within the LINSTOR system, using the
specified resource group. The name of the linstor resources is derived
from the IQN's World Wide Name, which must be unique'.
After that it creates a configuration for drbd-reactor to manage the
high availabilitiy primitives.

```
linstor-gateway iscsi create [flags]
```

### Examples

```
linstor-gateway iscsi create --iqn=iqn.2019-08.com.linbit:example --ip=192.168.122.181/24 --username=foo --lun=1 --password=bar --resource-group=ssd_thin_2way --size=2G
```

### Options

```
  -h, --help                    help for create
      --ip ip-cidr              Set the service IP and netmask of the target
  -i, --iqn iqn                 Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (default :)
  -l, --lun int                 Set the LUN (default 1)
  -p, --password string         Set the password (required)
      --portals string          Set up portals, if unset, the service ip and default port
  -g, --resource-group string   Set the LINSTOR resource-group (default "DfltRscGrp")
      --size unit               Set a size (e.g, 1TiB) (default 1GiB)
  -u, --username string         Set the username (required)
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

