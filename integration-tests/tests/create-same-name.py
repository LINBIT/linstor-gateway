#!/usr/bin/env python3
from subprocess import CalledProcessError

import gatewaytest

nodes = gatewaytest.setup()

first = nodes[0]
first.start_server()

first.run([
    'linstor-gateway', 'iscsi', 'create', 'iqn.2019-08.com.linbit:target1',
    '192.168.122.220/24', '1G'
])

try:
    first.run([
        'linstor-gateway', 'nvme', 'create', 'nqn.2021-08.com.linbit:nvme:target1',
        '192.168.122.222/24', '1G'
    ])
except CalledProcessError:
    print("command threw CalledProcessError, that was expected")
except BaseException:
    raise
finally:
    nodes.cleanup()
