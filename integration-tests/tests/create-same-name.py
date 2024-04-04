#!/usr/bin/env python3
from subprocess import CalledProcessError

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()
other_service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'iscsi', 'create', '--implementation=scst', 'iqn.2019-08.com.linbit:target1',
    service_ip, '1G'
])

try:
    first.run([
        'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:target1',
        other_service_ip, '1G'
    ])
except CalledProcessError:
    print("command threw CalledProcessError, that was expected")
except BaseException:
    raise

nodes.cleanup()
