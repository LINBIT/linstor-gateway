#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:nvme1',
    service_ip, '1G',
])
first.assert_resource_exists('nvme-of', 'nqn.2021-08.com.linbit:nvme:nvme1')

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active('nvme1')
gatewaytest.log('Resource nvme1 active on node {}'.format(active_node))
ls.wait_inuse_stable('nvme1', active_node)
gatewaytest.log('Resource nvme1 stably in use on node {}'.format(active_node))

first.run(['linstor-gateway', 'nvme', 'delete', '--force', 'nqn.2021-08.com.linbit:nvme:nvme1'])
first.assert_resource_not_exists('nvme-of', 'nqn.2021-08.com.linbit:nvme:nvme1')

assert not ls.resource_exists('nvme1')

ls.disconnect()
nodes.cleanup()
