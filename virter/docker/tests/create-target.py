#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]

print(first.run(['linstor', 'rg', 'create', 'group1']))
print(first.run(['linstor', 'vg', 'create', 'group1']))

print(first.run([
    'linstor-iscsi', 'create', '--iqn=iqn.2019-08.com.linbit:target1',
    '--lun=1', '--ip=192.168.122.222/24', '--username=admin', '--password=admin',
    '--resource-group=group1'
]))

