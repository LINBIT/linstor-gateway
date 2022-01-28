## linstor-gateway server

Starts a web server serving a REST API

### Synopsis

Starts a web server serving a REST API
An up to date version of the REST-API documentation can be found here:
https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway

For example:
linstor-gateway server --addr=":8080"

```
linstor-gateway server [flags]
```

### Options

```
      --addr string           Host and port as defined by http.ListenAndServe() (default ":8080")
      --controllers strings   List of LINSTOR controllers to try to connect to (default from $LS_CONTROLLERS, or localhost:3370)
  -h, --help                  help for server
```

### Options inherited from parent commands

```
      --config string     Config file to load (default "/etc/linstor-gateway/linstor-gateway.toml")
  -c, --connect string    LINSTOR Gateway server to connect to (default "http://localhost:8080")
      --loglevel string   Set the log level (as defined by logrus) (default "info")
```

### SEE ALSO

* [linstor-gateway](linstor-gateway.md)	 - Manage linstor-gateway targets and exports

