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
first.run(['linstor-gateway', 'nvme', 'delete', '--force', 'nqn.2021-08.com.linbit:nvme:nvme1'])
first.assert_resource_not_exists('nvme-of', 'nqn.2021-08.com.linbit:nvme:nvme1')

nodes.cleanup()
