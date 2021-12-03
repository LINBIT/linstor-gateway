#!/usr/bin/env python3
from subprocess import CalledProcessError

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]

first.run([
    'linstor-gateway', 'iscsi', 'create', '--iqn=iqn.2019-08.com.linbit:target1',
    '--lun=1', '--ip=192.168.122.220/24', '--username=admin', '--password=admin',
    '--resource-group=group1', '--size=1G'
])

try:
    first.run([
        'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:target1',
        '192.168.122.222/24', '1G', '--resource-group=group1'
    ])
except CalledProcessError:
    print("command threw CalledProcessError, that was expected")
except BaseException:
    raise
