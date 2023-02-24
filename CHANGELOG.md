# Changelog

## Unreleased

## 1.1.0 - 2023-02-24

### Features

* Extend the health check for NFS: it is now verified that the `nfs-server` package is actually installed.

### Miscellaneous

* Update Go dependencies

## 1.0.0 - 2022-11-21

* No changes over rc1

## 1.0.0-rc.1 - 2022-11-04

### Fixes

* Properly support CORS on the API endpoints
* Fix a nil dereference when one or more nodes are offline

### Features

* Add an `upgrade` command which can be used to migrate existing targets to newer
  versions of LINSTOR Gateway

### Miscellaneous

* Improve packaging process
* Update Go dependencies

## 0.13.1 - 2022-07-26

### Fixes

* Fix a bug that occurs when creating a volume with number zero

### Miscellaneous

* Update Go dependencies

## 0.13.0 - 2022-06-27

### Features

* Add a `--gross` option, which makes the "size" specify the actual space that the
  volume will occupy on disk instead of the usable net size

## 0.12.1 - 2022-05-03

### Fixes

* Make sure `iscsi create` respects the `--resource-group` argument
* Work around a size calculation bug for LINSTOR affecting thick LVM volumes

### Miscellaneous

* Update Go dependencies

## 0.12.0 - 2022-04-03

### Fixes

* Set the SCSI serial number to a value that is guaranteed to be (reasonably) unique

## 0.12.0-rc.1 - 2022-03-17

### Features

* Add a new endpoint, `/v2/status`, to query the status of the server
* Support CORS for the REST API
* Generate more readable toml files (`start` entries are on separate lines)
* Mark LINSTOR resources as degraded if the actual place count is lower than desired
* Point users to `linstor advise resource` if any resources are degraded

### Fixes

* Healthcheck: fix the path of the LINSTOR Satellite systemd override directory

### Miscellaneous

* Update Go dependencies

## 0.11.0 - 2022-02-14

* No changes over rc2

## 0.11.0-rc.2 - 2022-02-08

* Also add a cluster private volume for direct REST API calls
* Disallow underscores ("_") in iSCSI IQNs
* Add missing `--allowed-initiators` flag to iSCSI create command

## 0.11.0-rc.1 - 2022-01-31

* Add a new `check-health` command that checks whether all dependencies and requirements are met for LINSTOR Gateway
* Implement best practices for NFS and for failover scenarios
* Change architecture: the CLI is now exclusively a client for the server
* [REST API v2.0.0](https://app.swaggerhub.com/apis/Linstor/linstor-gateway/2.0.0)

## 0.10.0 - 2021-11-24

* No changes over rc1

## 0.10.0-rc.1 - 2021-11-17

* Change quorum options of created resource to be appropriate for drbd-reactor
* iSCSI: support supplying multiple service IPs for iSCSI-level multipathing
* Improvements to the documentation
* Change drbd-reactor configuration format to be compatible with drbd-reactor 0.5.0 (`on-stop-failure` ->
  `on-drbd-demote-failure`)

## 0.9.0 - 2021-09-28

* Minor fixes over rc3

## 0.9.0-rc.3 - 2021-09-23

* Minor fixes over rc2

## 0.9.0-rc.2 - 2021-09-15

* Some minor fixes over rc1
* The list of controllers is now parsed from a config file (additionally to the command line flag)

## 0.9.0-rc.1 - 2021-09-01

* Change high-availability backend from Pacemaker to LINBIT's own drbd-reactor
* Add support for NVME-oF targets
* Remove requirement for symlinking the binary: use "linstor-gateway iscsi" instead of "linstor-iscsi"
* Improve CI testing (integration tests)

## 0.8.0 - 2021-03-23

* Fix a bug related to Pacemaker 2 (use the "kind" attribute for order constraints instead of "score")
* Implement pipe detection; disable colors when not writing to a TTY
* Do not send back the iSCSI credentials over the REST API

## 0.7.0 - 2020-12-04

* Add ability to manage NFS exports
* Rename to linstor-gateway

## 0.6.2 - 2020-10-14

* Remove drbd-pacemaker depencency to fix packaging

## 0.6.1 - 2020-10-09

* Pacemaker now uses the drbd-attr resource agent to access drbd promotion scores

## 0.6.0 - 2020-04-29

* REST API version 1.1.0
* New command `linstor-iscsi version` displays version information
* New REST endpoints `/api/v1/iscsi/{iqn}/{lunid}/start` and `/stop` to start/stop targets
* `linstor-iscsi start`/`stop` now accepts the `--iqn` and `--lun` flags, making it actually work
* Fixes to the recent LUN size change

## 0.5.0 - 2020-03-31

* Size information was removed from LUN; it only belongs to LINSTOR

## 0.4.2 - 2020-03-09

* Improved process to find the LINSTOR controller
* Diskless LINSTOR resources are now considered "Good" instead of "Down"

## 0.4.1 - 2020-01-31

* Follow-up fixes to the service IP netmask problem

## 0.4.0 - 2020-01-31

* Fix service IP netmask (it was always 24 before)
* Status now includes information about what node the target is running on

## 0.3.1 - 2020-01-28

* Bump of gopacemaker

## 0.3.0 - 2020-01-20

* Service IP is now displayed when listing targets
* Low-level pacemaker API factored out to gopacemaker library

## 0.2.0 - 2019-12-03

* rpm packaging
* Bugfixes

## 0.1.0 - 2019-08-29

* First released version
* REST API version 1.0.0
