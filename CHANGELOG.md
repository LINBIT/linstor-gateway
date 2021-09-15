# Changelog

## Unreleased

## 0.9.0-rc.2 - 2021-09-15
* Some minor fixes over rc1
* The list of controllers is now parsed from a config file (additionally to the
  command line flag)

## 0.9.0-rc.1 - 2021-09-01
* Change high-availibility backend from Pacemaker to LINBIT's own drbd-reactor
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
