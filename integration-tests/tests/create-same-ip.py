#!/usr/bin/env python3
# create-same-ip
# - Create an iSCSI resource
# - Try to create an NVMe resource with the same IP address (should fail)
from subprocess import CalledProcessError

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'iscsi', 'create', 'iqn.2019-08.com.linbit:iscsi1',
    service_ip, '1G'
])

try:
    first.run([
        'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:nvme1',
        service_ip, '1G'
    ])
except CalledProcessError:
    print("command threw CalledProcessError, that was expected")
except BaseException:
    raise

nodes.cleanup()
