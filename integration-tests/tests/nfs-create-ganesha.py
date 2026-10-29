#! /usr/bin/env python3

import gatewaytest

# ----- HACK: ganesha-nfs RA is not in a released resource-agents yet -----
# Pull the RA from an upstream commit that supports generated mode
# (export_path, clients, ... as OCF parameters). Once a
# resource-agents release ships ocf:heartbeat:ganesha-nfs, drop this
# block and rely on the package install in the provisioning playbook.
GANESHA_NFS_RA_URL = (
    "https://raw.githubusercontent.com/ClusterLabs/resource-agents/"
    "d0e6423c5d5f0d949ef13bd4ae096f4e6ab0d88b/heartbeat/ganesha-nfs"
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
# Only allow the cluster's own network so the client below can mount.
allowed_ips = nodes.get_access_net()

first.run([
    'linstor-gateway', 'nfs', 'create', '--implementation=ganesha',
    '--allowed-ips', allowed_ips,
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
# Generated mode: linstor-gateway drives the export config through the RA's
# OCF parameters (allowed-ips -> clients, one export_path per volume,
# All_Squash to mirror the kernel implementation).
assert 'clients={}'.format(allowed_ips) in config_content, \
    'expected ganesha clients whitelist in promoter config, got:\n{}'.format(config_content)
assert 'export_path=/srv/gateway-exports/' in config_content, \
    'expected ganesha export_path in promoter config, got:\n{}'.format(config_content)
assert 'export_id=1' in config_content, \
    'expected ganesha export_id in promoter config, got:\n{}'.format(config_content)
assert 'squash=All_Squash' in config_content, \
    'expected ganesha squash policy in promoter config, got:\n{}'.format(config_content)

ls = gatewaytest.LinstorConnection(first)

active_node = ls.wait_for_resource_active('nfs1')
gatewaytest.log('Resource nfs1 active on node {}'.format(active_node))
ls.wait_inuse_stable('nfs1', active_node)
gatewaytest.log('Resource nfs1 stably in use on node {}'.format(active_node))

ip = service_ip.split('/')[0]
gatewaytest.check_failover(ls, nodes, 'nfs1', active_node,
                           connect=lambda client: client.nfs_mount(ip, '/srv/gateway-exports/nfs1'),
                           disconnect=lambda client: client.nfs_umount())

first.run(['linstor-gateway', 'nfs', 'delete', '--force', 'nfs1'])
first.assert_resource_not_exists('nfs', 'nfs1')

assert not ls.resource_exists('nfs1')

ls.disconnect()
nodes.cleanup()
