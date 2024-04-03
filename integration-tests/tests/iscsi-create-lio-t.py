#! /usr/bin/env python3

import gatewaytest

nodes = gatewaytest.setup()

nodes.run(['systemctl', 'stop', 'iscsi-scst'])  # special case for this test: avoid conflict with scst

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'iscsi', 'create', '--implementation=lio-t', 'iqn.2019-08.com.linbit:iscsi1',
    service_ip, '1G',
])
first.assert_resource_exists('iscsi', 'iqn.2019-08.com.linbit:iscsi1')
first.run(['linstor-gateway', 'iscsi', 'delete', '--force', 'iqn.2019-08.com.linbit:iscsi1'])
first.assert_resource_not_exists('iscsi', 'iqn.2019-08.com.linbit:iscsi1')

nodes.cleanup()
