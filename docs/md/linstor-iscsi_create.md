## linstor-iscsi create

Creates an iSCSI target

### Synopsis

Creates a highly available iSCSI target based on LINSTOR and Pacemaker.
At first it creates a new resouce within the linstor system, using the
specified resource group. The name of the linstor resources is derived
from the iqn and the lun number.
After that it creates resource primitives in the Pacemaker cluster including
all necessary order and location constraints. The Pacemaker primites are
prefixed with p_, contain the name and a resource type postfix.

For example:
linstor-iscsi create --iqn=iqn.2019-08.com.linbit:example --ip=192.168.122.181 \
 -username=foo --lun=1 --password=bar --resource_group=ssd_thin_2way --size=2G

Creates linstor resources example_lu0 and
pacemaker primitives p_iscsi_example_ip, p_iscsi_example, p_iscsi_example_lu0

```
linstor-iscsi create [flags]
```

### Options

```
  -c, --controller ip           Set the IP of the linstor controller node (default 127.0.0.1)
  -h, --help                    help for create
      --ip ip                   Set the service IP of the target (required) (default 127.0.0.1)
  -p, --password string         Set the password (required)
      --portals string          Set up portals, if unset, the service ip and default port
  -g, --resource-group string   Set the LINSTOR resource-group (default "default")
      --size unit               Set a size (e.g, 1TiB) (default 1GiB)
  -u, --username string         Set the username (required)
```

### Options inherited from parent commands

```
  -i, --iqn string        Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)
      --loglevel string   Set the log level (as defined by logrus) (default "info")
  -l, --lun int           Set the LUN Number (required) (default 1)
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

