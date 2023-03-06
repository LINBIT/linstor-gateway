#! /usr/bin/env python3
# iscsi-dirty-delete
# - Create an iSCSI resource
# - Delete its LINSTOR resource
# - Try to create the same resource again
#
# The existing config should be ignored and the resource should be created
# as normal.
import time

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()

first.run([
    'linstor-gateway', 'iscsi', 'create', 'iqn.2019-08.com.linbit:iscsi1',
    '192.168.122.230/24', '1G',
])

nodes.run(['systemctl', 'stop', 'drbd-reactor'])
time.sleep(5)  # TODO find a different solution
first.run(['linstor', 'resource-definition', 'delete', 'iscsi1'])
nodes.run(['systemctl', 'start', 'drbd-reactor'])

first.run([
    'linstor-gateway', 'iscsi', 'create', 'iqn.2019-08.com.linbit:iscsi1',
    '192.168.122.230/24', '1G',
])

nodes.cleanup()
