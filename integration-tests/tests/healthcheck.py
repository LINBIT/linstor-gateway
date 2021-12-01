#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
print(first.run(['linstor-gateway', 'check-health']))
