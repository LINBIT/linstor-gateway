#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()

first.run([
    'linstor-gateway', 'iscsi', 'create', 'iqn.2019-08.com.linbit:iscsi1',
    '192.168.122.220/24', '1G',
])
first.run(['linstor-gateway', 'iscsi', 'delete', 'iqn.2019-08.com.linbit:iscsi1'])

nodes.cleanup()
