# Changelog

## Unreleased
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
