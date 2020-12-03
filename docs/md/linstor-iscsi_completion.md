## linstor-iscsi completion

Generates bash completion script

### Synopsis

To load completion run

. <(linstor-iscsi completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(linstor-iscsi completion)

```
linstor-iscsi completion [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-iscsi](linstor-iscsi.md)	 - Manages Highly-Available iSCSI targets

