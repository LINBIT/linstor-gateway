## linstor-gateway iscsi upgrade

Check existing resources and upgrade their configuration if necessary

```
linstor-gateway iscsi upgrade IQN [flags]
```

### Options

```
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
  -d, --dry-run               Display potential updates without taking any actions
  -h, --help                  help for upgrade
  -y, --yes                   Run non-interactively; answer all questions with yes
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway iscsi](linstor-gateway_iscsi.md)	 - Manages Highly-Available iSCSI targets

