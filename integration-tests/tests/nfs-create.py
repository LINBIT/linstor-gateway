#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run(['linstor-gateway', 'nfs', 'create', 'nfs1', service_ip, '1G'])
first.assert_resource_exists('nfs', 'nfs1')

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active('nfs1')
gatewaytest.log('Resource nfs1 active on node {}'.format(active_node))
ls.wait_inuse_stable('nfs1', active_node)
gatewaytest.log('Resource nfs1 stably in use on node {}'.format(active_node))

first.run(['linstor-gateway', 'nfs', 'delete', '--force', 'nfs1'])
first.assert_resource_not_exists('nfs', 'nfs1')

assert not ls.resource_exists('nfs1')

ls.disconnect()
nodes.cleanup()
