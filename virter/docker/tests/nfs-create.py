#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]

print(first.run(['linstor', 'rg', 'create', 'group1']))
print(first.run(['linstor', 'vg', 'create', 'group1']))

print(first.run([
    'linstor-gateway', 'nfs', 'create', '--resource=nfs1', '--service-ip=192.168.122.221/24',
    '--allowed-ips=192.168.0.0/16', '--resource-group=group1', '--size=1G'
]))

print(first.run([
    'linstor-gateway', 'nfs', 'delete', '--resource=nfs1',
]))
