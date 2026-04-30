#! /usr/bin/env python3

# Integration test for `linstor-gateway nvme add-volume` and
# `nvme delete-volume`. NVMe-oF mirrors iSCSI's volume layout:
# a "cluster private" system volume at index 0, then user namespaces
# numbered from 1. Both add-volume and delete-volume require the
# target to be stopped, guarded by:
#   pkg/nvmeof/nvmeof.go: "cannot delete volume while service is running"
#
# The test creates a target with one namespace (NSID 1), adds NSID 2,
# verifies all three LINSTOR volumes (system + 1 + 2) reach UpToDate,
# then deletes the *non-trailing* namespace (NSID 1) and asserts the
# remaining sparse {0, 2} layout is healthy and the target still
# serves. Also asserts that delete-volume refuses while running.

from subprocess import CalledProcessError

import gatewaytest

NQN = 'nqn.2021-08.com.linbit:nvme:nvme1'
RESOURCE = 'nvme1'

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

# --- Build a target with namespace 1. ---------------------------------
first.run([
    'linstor-gateway', 'nvme', 'create', NQN, service_ip, '1G',
])
first.assert_resource_exists('nvme-of', NQN)

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active(RESOURCE)
gatewaytest.log('Resource {} active on node {}'.format(RESOURCE, active_node))
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Negative case: delete-volume must refuse while running. ----------
try:
    first.run(['linstor-gateway', 'nvme', 'delete-volume', NQN, '1'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly refused while target is running')
else:
    raise AssertionError('delete-volume should have failed while target was running')

# --- Stop, add namespace 2, restart, confirm it serves. ---------------
# nvme add-volume requires an explicit volume number (no auto-increment).
first.run(['linstor-gateway', 'nvme', 'stop', NQN])
first.run(['linstor-gateway', 'nvme', 'add-volume', NQN, '2', '1G'])

# system volume + NSID 1 + NSID 2 = 3
ls.wait_all_volumes_uptodate(RESOURCE, expected_volumes=3)
gatewaytest.log('All 3 volumes UpToDate after add-volume')

first.run(['linstor-gateway', 'nvme', 'start', NQN])
active_node = ls.wait_for_resource_active(RESOURCE)
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Stop, delete NSID 1 (the non-trailing one), restart, verify. -----
first.run(['linstor-gateway', 'nvme', 'stop', NQN])
first.run(['linstor-gateway', 'nvme', 'delete-volume', NQN, '1'])

# system volume + NSID 2 = 2; volume 1 must be gone on every node.
ls.wait_all_volumes_uptodate(RESOURCE, expected_volumes=2)
gatewaytest.log('All remaining volumes UpToDate after delete-volume')

first.run(['linstor-gateway', 'nvme', 'start', NQN])
active_node = ls.wait_for_resource_active(RESOURCE)
gatewaytest.log('Resource {} active again on node {}'.format(RESOURCE, active_node))
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Probe: deleting a non-existent NSID must fail loudly. ------------
# Used to silently report success and do nothing; now expected to error
# (pkg/nvmeof/nvmeof.go DeleteVolume returns "volume N does not exist").
first.run(['linstor-gateway', 'nvme', 'stop', NQN])
try:
    first.run(['linstor-gateway', 'nvme', 'delete-volume', NQN, '99'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly errored for a non-existent NSID')
else:
    raise AssertionError(
        'delete-volume should have errored for a non-existent NSID')

# --- Probe: deleting the last remaining namespace must be refused. ----
# A target with zero user namespaces is a useless half-state -- the
# right way to remove the last NSID is `nvme delete`.
try:
    first.run(['linstor-gateway', 'nvme', 'delete-volume', NQN, '2'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly refused for the last user namespace')
else:
    raise AssertionError(
        'delete-volume should have refused removing the last remaining namespace')

# --- Cleanup. ---------------------------------------------------------
first.run(['linstor-gateway', 'nvme', 'delete', '--force', NQN])
first.assert_resource_not_exists('nvme-of', NQN)

assert not ls.resource_exists(RESOURCE)

ls.disconnect()
nodes.cleanup()
