#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]

print(first.run(['linstor', 'rg', 'create', 'group1']))
print(first.run(['linstor', 'vg', 'create', 'group1']))

print(first.run([
    'linstor-gateway', 'iscsi', 'create', '--iqn=iqn.2019-08.com.linbit:iscsi1',
    '--lun=1', '--ip=192.168.122.220/24', '--username=admin', '--password=admin',
    '--resource-group=group1', '--size=1G'
]))

print(first.run([
    'linstor-gateway', 'iscsi', 'delete', '--iqn=iqn.2019-08.com.linbit:iscsi1',
]))
