## linstor-gateway completion

Generates bash completion script

### Synopsis

To load completion run

. <(linstor-gateway completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(linstor-gateway completion)

```
linstor-gateway completion [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8337")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports

