# linstor-iscsi

`linstor-iscsi` manages highly available iSCSI targets by leveraging on LINSTOR
and Pacemaker. Setting up LINSTOR - including storage pools and resource groups -
as well as Corosync and Pacemaker's properties are a prerequisite to use this tool.

# Building
Use a version of go that supports modules (>1.11). Then you can `go get` the code as usual.

```
go get github.com/LINBIT/linstor-iscsi
```

# Documentation
Start by browsing the documentation [here](./docs/md/linstor-iscsi.md).

The REST-API documentation can be found [here](https://app.swaggerhub.com/apis-docs/Linstor/linstor-iscsi/).
