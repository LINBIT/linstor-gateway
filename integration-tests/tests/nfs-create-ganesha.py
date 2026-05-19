#! /usr/bin/env python3

import gatewaytest

# ----- HACK: ganesha-nfs RA is not in a released resource-agents yet -----
# Pull the RA from the upstream commit that introduced it. Once a
# resource-agents release ships ocf:heartbeat:ganesha-nfs, drop this
# block and rely on the package install in the provisioning playbook.
GANESHA_NFS_RA_URL = (
    "https://raw.githubusercontent.com/ClusterLabs/resource-agents/"
    "7c15379357a469c8060b13ec8afaed13d667f2e4/heartbeat/ganesha-nfs"
)
GANESHA_NFS_RA_PATH = "/usr/lib/ocf/resource.d/heartbeat/ganesha-nfs"
# ------------------------------------------------------------------------

nodes = gatewaytest.setup()

for n in nodes:
    n.run(['curl', '-fsSL', '-o', GANESHA_NFS_RA_PATH, GANESHA_NFS_RA_URL])
    n.run(['chmod', '+x', GANESHA_NFS_RA_PATH])

first = nodes[0]
first.start_server()
service_ip = nodes.get_service_ip()

first.run([
    'linstor-gateway', 'nfs', 'create', '--implementation=ganesha',
    'nfs1', service_ip, '1G',
])
first.assert_resource_exists('nfs', 'nfs1')

# Verify the generated promoter config uses the ganesha-nfs RA and has
# dropped the kernel-NFS-specific agents.
config_content = first.run(
    ['cat', '/etc/drbd-reactor.d/linstor-gateway-nfs-nfs1.toml'],
    return_stdout=True,
)
assert 'ocf:heartbeat:ganesha-nfs' in config_content, \
    'expected ocf:heartbeat:ganesha-nfs in promoter config, got:\n{}'.format(config_content)
assert 'ocf:heartbeat:nfsserver' not in config_content, \
    'unexpected ocf:heartbeat:nfsserver in ganesha config:\n{}'.format(config_content)
assert 'ocf:heartbeat:exportfs' not in config_content, \
    'unexpected ocf:heartbeat:exportfs in ganesha config:\n{}'.format(config_content)

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
