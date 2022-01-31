#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()

first.run([
    'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:nvme1',
    '192.168.122.222/24', '1G',
])
first.run(['linstor-gateway', 'nvme', 'delete', 'nqn.2021-08.com.linbit:nvme:nvme1'])

nodes.cleanup()
