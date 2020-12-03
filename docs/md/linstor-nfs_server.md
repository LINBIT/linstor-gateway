## linstor-nfs server

Starts a web server serving a REST API

### Synopsis

Starts a web server serving a REST API
An up to date version of the REST-API documentation can be found here:
https://app.swaggerhub.com/apis-docs/Linstor/linstor-iscsi/

For example:
linstor-iscsi server --addr=":8080"

```
linstor-nfs server [flags]
```

### Options

```
      --addr string   Host and port as defined by http.ListenAndServe() (default ":8080")
  -h, --help          help for server
```

### Options inherited from parent commands

```
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-nfs](linstor-nfs.md)	 - Manages Highly-Available NFS exports

