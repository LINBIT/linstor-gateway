## linstor-gateway nfs server

Starts a web server serving a REST API

### Synopsis

Starts a web server serving a REST API
An up to date version of the REST-API documentation can be found here:
https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway

For example:
linstor-gateway server --addr=":8080"

```
linstor-gateway nfs server [flags]
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

* [linstor-gateway nfs](linstor-gateway_nfs.md)	 - Manages Highly-Available NFS exports

