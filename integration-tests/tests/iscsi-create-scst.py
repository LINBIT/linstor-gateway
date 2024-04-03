#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'iscsi', 'create', '--implementation=scst', 'iqn.2019-08.com.linbit:iscsi1',
    service_ip, '1G',
])
first.assert_resource_exists('iscsi', 'iqn.2019-08.com.linbit:iscsi1')

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active('iscsi1')
gatewaytest.log('Resource iscsi1 active on node {}'.format(active_node))
ls.wait_inuse_stable('iscsi1', active_node)
gatewaytest.log('Resource iscsi1 stably in use on node {}'.format(active_node))

first.run(['linstor-gateway', 'iscsi', 'delete', '--force', 'iqn.2019-08.com.linbit:iscsi1'])
first.assert_resource_not_exists('iscsi', 'iqn.2019-08.com.linbit:iscsi1')

assert not ls.resource_exists('iscsi1')

nodes.cleanup()
