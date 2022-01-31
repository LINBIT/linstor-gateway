#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()

first.run(['linstor-gateway', 'nfs', 'create', 'nfs1', '192.168.122.221/24', '1G'])
first.run(['linstor-gateway', 'nfs', 'delete', 'nfs1'])

nodes.cleanup()
