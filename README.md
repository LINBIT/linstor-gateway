<p align="center"><a href="https://linbit.com/linstor"><img src="https://raw.githubusercontent.com/LINBIT/linstor-gateway/master/docs/Linstor-Logo.png" width="400" alt="LINSTOR Logo"/></a></p>

# LINSTOR Gateway

<a href="https://github.com/LINBIT/linstor-gateway/releases"><img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/LINBIT/linstor-gateway"></a> <a href="https://github.com/LINBIT/linstor-gateway/blob/master/LICENSE"><img alt="GitHub" src="https://img.shields.io/github/license/LINBIT/linstor-gateway"></a> <a href="https://github.com/LINBIT/linstor-gateway/actions"><img alt="GitHub Workflow Status" src="https://img.shields.io/github/actions/workflow/status/LINBIT/linstor-gateway/go.yml"></a> <a href="https://join.slack.com/t/linbit-community/shared_invite/enQtOTg0MTEzOTA4ODY0LTFkZGY3ZjgzYjEzZmM2OGVmODJlMWI2MjlhMTg3M2UyOGFiOWMxMmI1MWM4Yjc0YzQzYWU0MjAzNGRmM2M5Y2Q"><img alt="Slack Channel" src="https://img.shields.io/badge/Slack-linbit--community-green"/></a>

LINSTOR Gateway manages highly available **iSCSI targets**, **NFS exports**, and
**NVMe-oF targets** by leveraging [LINSTOR](https://github.com/LINBIT/linstor-server)
and [drbd-reactor](https://github.com/LINBIT/drbd-reactor).

# Quick Start

1. Set up a [LINSTOR](https://github.com/LINBIT/linstor-server) cluster. Ensure you have
   a [storage pool](https://linbit.com/drbd-user-guide/linstor-guide-1_0-en/#s-storage_pools) with enough space for your
   data.
2. Set up [drbd-reactor](https://github.com/LINBIT/drbd-reactor). The daemon should be configured to reload
   automatically when the configuration changes – follow the steps from
   the [drbd-reactor README](https://github.com/LINBIT/drbd-reactor/#automatic-reload).
3. LINSTOR Gateway is packaged as a single binary. Download one of the
   [releases](https://github.com/LINBIT/linstor-gateway/releases) and copy it to `/usr/local/bin`.
4. Start the LINSTOR Gateway server:
   ```
   linstor-gateway server
   ```
   This repository also includes [linstor-gateway.service](./linstor-gateway.service)
   which can be used to start the server via systemd.

# Requirements

LINSTOR Gateway provides a built-in health check that automatically tests whether all requirements are correctly met on
the current host.

Simply execute

```
linstor-gateway check-health
```

and follow any suggestions that may come up.

# Documentation

If you want to learn more about LINSTOR Gateway, here are some pointers for further reading.

### Command Line

Help for the command line interface is available by running:

```
linstor-gateway help
```

The same information can also be browsed in Markdown format [here](./docs/md/linstor-gateway.md).

### Configuration

LINSTOR Gateway takes a configuration file. See its documentation [here](./docs/config.md).

### Internals

The LINSTOR Gateway command line client communicates with the server by using a REST API, which is
documented [here](https://app.swaggerhub.com/apis-docs/Linstor/linstor-gateway/).

It also exposes a Go client for the REST
API: <a href="https://pkg.go.dev/github.com/LINBIT/linstor-gateway/client"><img src="https://pkg.go.dev/badge/github.com/LINBIT/linstor-gateway/client.svg" alt="Go Reference"></a>

# Building

If you want to test the latest unstable version of LINSTOR Gateway, you can build the git version from sources:

```
git clone https://github.com/LINBIT/linstor-gateway
cd linstor-gateway
make
```
