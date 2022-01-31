#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
first.run(['linstor-gateway', 'check-health'])

nodes.cleanup()
