#! /usr/bin/env python3

# Integration test for `linstor-gateway iscsi add-volume` and
# `iscsi delete-volume`.
#
# Combines what used to be two tests. The add-volume half is the
# original regression test for linstor/linstor-gateway#24
# ("newly added volume Inconsistent -> target cannot start") --
# we still create with one LUN, add a second, and assert all three
# LINSTOR volumes (system + LUN 1 + LUN 2) reach UpToDate before
# starting the target back up. The delete-volume half then removes
# the *non-trailing* LUN 1 -- exercising volume bookkeeping more
# thoroughly than removing the most-recently-added one -- and
# asserts the surviving sparse {0, 2} layout is healthy and still
# serves. Also asserts that delete-volume refuses while the target
# is started, matching the "cannot delete volume while service is
# running" guard in pkg/iscsi/iscsi.go.

from subprocess import CalledProcessError

import gatewaytest

IQN = 'iqn.2019-08.com.linbit:iscsi1'
RESOURCE = 'iscsi1'

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

# --- Build a target with two user LUNs: 1 and 2. -----------------------
first.run([
    'linstor-gateway', 'iscsi', 'create', '--implementation=scst', IQN,
    service_ip, '1G',
])
first.assert_resource_exists('iscsi', IQN)

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active(RESOURCE)
gatewaytest.log('Resource {} active on node {}'.format(RESOURCE, active_node))
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Negative case: delete-volume must refuse while running. ----------
# Doing this before stop avoids the cost of an extra stop/start cycle.
try:
    first.run(['linstor-gateway', 'iscsi', 'delete-volume', IQN, '1'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly refused while target is running')
else:
    raise AssertionError('delete-volume should have failed while target was running')

# --- Stop, add a second LUN, restart, confirm it serves. --------------
first.run(['linstor-gateway', 'iscsi', 'stop', IQN])
first.run(['linstor-gateway', 'iscsi', 'add-volume', IQN, '1G'])

# system volume + LUN 1 + LUN 2 = 3
ls.wait_all_volumes_uptodate(RESOURCE, expected_volumes=3)
gatewaytest.log('All 3 volumes UpToDate after add-volume')

first.run(['linstor-gateway', 'iscsi', 'start', IQN])
active_node = ls.wait_for_resource_active(RESOURCE)
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Stop, delete LUN 1 (the non-trailing one), restart, verify. ------
first.run(['linstor-gateway', 'iscsi', 'stop', IQN])
first.run(['linstor-gateway', 'iscsi', 'delete-volume', IQN, '1'])

# system volume + LUN 2 = 2; volume 1 must be gone on every node.
ls.wait_all_volumes_uptodate(RESOURCE, expected_volumes=2)
gatewaytest.log('All remaining volumes UpToDate after delete-volume')

first.run(['linstor-gateway', 'iscsi', 'start', IQN])
active_node = ls.wait_for_resource_active(RESOURCE)
gatewaytest.log('Resource {} active again on node {}'.format(RESOURCE, active_node))
ls.wait_inuse_stable(RESOURCE, active_node)

# --- Probe: deleting a non-existent LUN must fail loudly. -------------
# Used to silently report success and do nothing; now expected to error
# (pkg/iscsi/iscsi.go DeleteVolume returns "volume N does not exist").
first.run(['linstor-gateway', 'iscsi', 'stop', IQN])
try:
    first.run(['linstor-gateway', 'iscsi', 'delete-volume', IQN, '99'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly errored for a non-existent LUN')
else:
    raise AssertionError(
        'delete-volume should have errored for a non-existent LUN')

# --- Probe: deleting the last remaining user LUN must be refused. -----
# A target with zero user LUNs is a useless half-state -- the right way
# to remove the last LUN is `iscsi delete`.
try:
    first.run(['linstor-gateway', 'iscsi', 'delete-volume', IQN, '2'])
except CalledProcessError:
    gatewaytest.log('delete-volume correctly refused for the last user LUN')
else:
    raise AssertionError(
        'delete-volume should have refused removing the last remaining user LUN')

# --- Cleanup. ---------------------------------------------------------
first.run(['linstor-gateway', 'iscsi', 'delete', '--force', IQN])
first.assert_resource_not_exists('iscsi', IQN)

assert not ls.resource_exists(RESOURCE)

ls.disconnect()
nodes.cleanup()
