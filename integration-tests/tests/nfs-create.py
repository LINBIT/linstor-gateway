#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run(['linstor-gateway', 'nfs', 'create', 'nfs1', service_ip, '1G'])
first.assert_resource_exists('nfs', 'nfs1')
first.run(['linstor-gateway', 'nfs', 'delete', '--force', 'nfs1'])
first.assert_resource_not_exists('nfs', 'nfs1')

nodes.cleanup()
