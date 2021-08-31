#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]

print(first.run(['linstor', 'rg', 'create', 'group1']))
print(first.run(['linstor', 'vg', 'create', 'group1']))

print(first.run([
    'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:nvme1',
    '192.168.122.222/24', '1G', '--resource-group=group1'
]))

print(first.run([
    'linstor-gateway', 'nvme', 'delete', 'nqn.2021-08.com.linbit:nvme:nvme1'
]))
print('test')
