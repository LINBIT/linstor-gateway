#! /usr/bin/env python3

import argparse
import errno
import ipaddress
import os
import re
import shlex
import socket
import subprocess
import sys
import time
from contextlib import contextmanager
from io import StringIO
from threading import Thread

import linstor
from linstor.errors import LinstorNetworkError
import requests
from lbpytest.controlmaster import SSH

# stream to write output to
logstream = None

# client-side paths used by Node.start_io() and friends
IO_MOUNTPOINT = '/mnt/gatewaytest'
IO_PATTERN = '/tmp/gatewaytest-pattern'
IO_COUNT = '/tmp/gatewaytest-io.count'
IO_FAILED = '/tmp/gatewaytest-io.failed'

# save the original excepthook, see handle_exception() below
__real_excepthook = sys.excepthook

# map the resource type to the name of the identifier field in the API
id_map = {
    'nfs': 'name',
    'nvme-of': 'nqn',
    'iscsi': 'iqn',
}


class Tee(object):
    """
    replicates writes to streams
    """

    def __init__(self):
        self.streams = set()

    def add(self, stream):
        self.streams.add(stream)

    def remove(self, stream):
        self.streams.remove(stream)

    def write(self, message):
        for stream in self.streams:
            stream.write(message)

    def flush(self):
        for stream in self.streams:
            stream.flush()


def log(*args, **kwargs):
    """ Print message to stderr """
    print(*args, file=logstream)
    logstream.flush()


class LinstorConnection(linstor.Linstor):
    def __init__(self, node):
        super(LinstorConnection, self).__init__("linstor://" + node.addr)
        self.connect()

    def __del__(self):
        self.disconnect()

    def wait_for_resource_active(self, resource: str, retries: int = 5):
        """
        Wait for a resource to become active. Active is defined as having only Diskless or UpToDate resources and at
        least 2 UpToDate resources, while also having the resource InUse on one node.

        :returns: the node name on which the resource is InUse
        :exception: TimeoutError: if the resource does not reach the desired state after the specified number of retries
        """

        want_states = ['Diskless', 'UpToDate']
        attempts = 0

        def do_attempt():
            resp = self.resource_list(filter_by_resources=[resource])
            if len(resp) == 0:
                return False
            states = resp[0].resource_states
            if len(states) == 0:
                return False

            num_uptodate = 0
            num_diskless = 0
            inuse = ''
            for st in states:
                if st.volume_states[0].disk_state == 'UpToDate':
                    num_uptodate += 1
                if st.volume_states[0].disk_state == 'Diskless':
                    num_diskless += 1
                if st.in_use:
                    inuse = st.node_name

            if num_uptodate + num_diskless == len(states) and num_uptodate >= 2 and inuse != '':
                return inuse
            print('Resource {} not yet active, UpToDate: {} / Diskless: {} / InUse: {}'.format(resource, num_uptodate,
                                                                                               num_diskless, inuse))
            return False

        while True:
            attempts += 1
            if attempts > retries:
                raise TimeoutError(
                    'Resource {} did not reach state {} after {} retries'.format(resource, want_states, retries))

            inuse_node = do_attempt()
            if inuse_node:
                return inuse_node

            time.sleep(1)

    def wait_inuse_stable(self, resource: str, node: str, count_required: int = 5):
        """
        Wait for a resource to be InUse on the specified node for a certain number of attempts.
        """

        count = 0

        def do_attempt():
            resp = self.resource_list(filter_by_resources=[resource])
            if len(resp) == 0:
                raise Exception('Resource {} not found'.format(resource))
            states = resp[0].resource_states
            if len(states) == 0:
                raise Exception('Resource {} has no states'.format(resource))

            for st in states:
                if st.in_use:
                    if st.node_name == node:
                        return True
                    raise Exception(
                        'Resource {} is in use by node {}, expected {}'.format(resource, st.node_name, node))
            raise Exception('Resource {} is not in use on any node'.format(resource))

        while True:
            count += 1
            if count > count_required:
                return True
            do_attempt()

            time.sleep(1)

    def wait_all_volumes_uptodate(self, resource: str, expected_volumes: int,
                                  retries: int = 30, sleep_s: float = 2.0):
        """
        Wait until every volume of the resource is UpToDate on all diskful
        nodes (and Diskless on diskless nodes). Each volume must have at
        least 2 UpToDate replicas. Catches the issue #24 condition where a
        newly added volume stays Inconsistent on the diskful peers.

        :param resource: the LINSTOR resource name (e.g. "iscsi1")
        :param expected_volumes: number of volumes the resource must have
        :param retries: number of polling attempts before giving up
        :param sleep_s: delay between attempts
        :exception TimeoutError: if the condition is not reached in time
        """

        attempts = 0
        while True:
            attempts += 1
            if attempts > retries:
                raise TimeoutError(
                    'Resource {} did not have all {} volumes UpToDate after {} retries'.format(
                        resource, expected_volumes, retries))

            resp = self.resource_list(filter_by_resources=[resource])
            if len(resp) == 0 or len(resp[0].resource_states) == 0:
                time.sleep(sleep_s)
                continue
            states = resp[0].resource_states

            ok = True
            bad = []
            # Key by LINSTOR volume number, not list index: after a
            # non-trailing delete-volume the surviving volumes can be
            # sparse (e.g. {0, 2}), so a positional list would alias
            # them onto the same slot.
            per_volume_uptodate = {}
            for st in states:
                if len(st.volume_states) != expected_volumes:
                    ok = False
                    bad.append('node={} has {} volumes (expected {})'.format(
                        st.node_name, len(st.volume_states), expected_volumes))
                    continue
                for v in st.volume_states:
                    per_volume_uptodate.setdefault(v.number, 0)
                    if v.disk_state == 'UpToDate':
                        per_volume_uptodate[v.number] += 1
                    elif v.disk_state == 'Diskless':
                        pass
                    else:
                        ok = False
                        bad.append('node={} vol={} state={}'.format(
                            st.node_name, v.number, v.disk_state))

            if ok and len(per_volume_uptodate) == expected_volumes \
                    and all(c >= 2 for c in per_volume_uptodate.values()):
                return True

            log('Resource {} not yet healthy (attempt {}/{}): uptodate-per-volume={} bad={}'.format(
                resource, attempts, retries, per_volume_uptodate, bad))
            time.sleep(sleep_s)

    def resource_list(self, *args, **kwargs):
        # The controller closes idle keep-alive connections, e.g. while a test
        # is busy with a client; reconnect once instead of failing.
        try:
            return super().resource_list(*args, **kwargs)
        except LinstorNetworkError:
            log('LINSTOR connection lost, reconnecting')
            self.disconnect()
            self.connect()
            return super().resource_list(*args, **kwargs)

    def wait_for_failover(self, resource: str, old_node: str, retries: int = 120) -> str:
        """
        Wait for a resource to become InUse on a node other than old_node, and
        for old_node to have released it.

        :returns: the node name on which the resource is now InUse
        :exception: TimeoutError: if the resource does not move in time
        """
        for _ in range(retries):
            resp = self.resource_list(filter_by_resources=[resource])
            states = resp[0].resource_states if resp else []
            in_use = [st.node_name for st in states if st.in_use]
            if len(in_use) == 1 and in_use[0] != old_node:
                return in_use[0]
            time.sleep(1)
        raise TimeoutError('Resource {} did not fail over from {} after {} retries'.format(
            resource, old_node, retries))

    def resource_exists(self, resource: str):
        """
        Check if a resource exists.
        """
        resp = self.resource_list(filter_by_resources=[resource])
        return len(resp) > 0 and len(resp[0].resource_states) > 0


class Nodes(list):
    def __init__(self, members=[]):
        super(Nodes, self).__init__(members)
        self.access_net = os.getenv('VIRTER_ACCESS_NETWORK')

    def run(self, command):
        return [n.run(command) for n in self]

    def cleanup(self):
        log('Cleaning up nodes')
        for n in self:
            n.cleanup()

    def by_name(self, name: str) -> 'Node':
        """
        :param name: a LINSTOR node name
        """
        for n in self:
            if n.is_named(name):
                return n
        raise RuntimeError('Node {} not found'.format(name))

    def client_for(self, *busy_nodes: str) -> 'Node':
        """
        Pick a node to act as a client, preferring one that does not host the
        resource so traffic actually goes through the service IP.

        :param busy_nodes: LINSTOR node names that must not be used as client
        """
        for n in self:
            if not any(n.is_named(b) for b in busy_nodes):
                return n
        raise RuntimeError('No node other than {} available as client'.format(', '.join(busy_nodes)))

    def get_service_ip(self) -> str:
        """
        Get an available IP address from the access network, including the
        network mask of the access network.

        :return: an IP address with network mask as a string (e.g. '192.168.1.254/24')
        """

        if self.access_net is None:
            raise RuntimeError('No access network defined')
        return '{}/{}'.format(self.get_available_ip(), self.get_access_net_mask())

    def get_available_ip(self) -> str:
        """
        Get an IP address from the access network that is not used by any node
        in the cluster. The search is started from the end of the network to
        maximize the chance of getting an actually unused IP address.

        :return: an IP address as a string (e.g. '192.168.1.254')
        """
        if self.access_net is None:
            raise RuntimeError('No access network defined')
        net = ipaddress.ip_network(self.access_net, strict=False)
        addrs = [a for a in net.hosts()]
        for n in self:
            if n.addr in addrs:
                addrs.remove(n.addr)
        for a in reversed(addrs):
            return str(a)
        raise RuntimeError('No available IP address found')

    def get_access_net(self) -> str:
        """
        :return: the access network in CIDR notation (e.g. '192.168.1.0/24')
        """
        if self.access_net is None:
            raise RuntimeError('No access network defined')
        return str(ipaddress.ip_network(self.access_net, strict=False))

    def get_access_net_mask(self) -> int:
        if self.access_net is None:
            raise RuntimeError('No access network defined')
        net = ipaddress.ip_network(self.access_net, strict=False)
        return net.prefixlen


class Node:
    def __init__(self, name, logdir, addr=None):
        self.name = name
        try:
            self.addr = addr if addr else socket.gethostbyname(name)
        except:
            raise RuntimeError('Could not determine IP for host %s' % name)
        self.ssh = SSH(self.addr)
        self.hostname = self.run(['hostname', '-f'], return_stdout=True)
        self.server_process = None
        self.logdir = logdir
        self.io_pid = None

    def is_named(self, name: str) -> bool:
        return name in (self.name, self.hostname, self.hostname.split('.')[0])

    def save_logs(self):
        log('Saving logs for {}'.format(self.name))
        journal_log = self.run(["journalctl"], return_stdout=True)
        with open(os.path.join(self.logdir, '{}-journal.log'.format(self.name)), 'w') as f:
            f.write(journal_log)

        dmesg_log = self.run(["dmesg"], return_stdout=True)
        with open(os.path.join(self.logdir, '{}-dmesg.log'.format(self.name)), 'w') as f:
            f.write(dmesg_log)

        lsmod_log = self.run(["lsmod"], return_stdout=True)
        with open(os.path.join(self.logdir, '{}-lsmod.log'.format(self.name)), 'w') as f:
            f.write(lsmod_log)

    def cleanup(self):
        self.stop_server()
        self.save_logs()
        self.ssh.close()

    def start_server(self):
        def server(n: Node):
            p = n.ssh.Popen('linstor-gateway --loglevel=trace server')
            n.server_process = p
            n.ssh.pipeIO(p, stdout=logstream, stderr=logstream)

        thread = Thread(target=server, args=(self,))
        thread.start()
        # sleep for 1 second so that the server is definitely ready
        # TODO: literally any other solution would be better, so think of one...
        time.sleep(1)

    def stop_server(self):
        if self.server_process:
            log('Stopping server on {}'.format(self.name))
            self.server_process.terminate()
            self.server_process = None

    def run(self, cmd, quote=True, catch=False, return_stdout=False, stdin=None, stdout=None,
            stderr=None, env=None):
        """
        Run a command via SSH on the target node.

        :param cmd: the command as a list of strings
        :param quote: use shell quoting to prevent environment variable substitution in commands
        :param catch: report command failures on stderr rather than raising an exception
        :param return_stdout: return the stdout returned by the command instead of printing it
        :param stdin: standard input to command (file-like object)
        :param stdout: standard output from command (file-like object)
        :param stderr: standard error from command (file-like object)
        :param env: a dictionary of extra environment variables which will be exported to the command
        :returns: nothing, or a string if return_stdout is True
        :raise CalledProcessError: when the command fails (unless catch is True)
        """
        stdout = stdout or logstream
        stderr = stderr or logstream
        stdin = stdin or False  # False means no stdin
        if return_stdout:
            # if stdout should be returned, do not log stdout to logstream too
            stdout = StringIO()

        if quote:
            cmd_string = ' '.join(shlex.quote(str(x)) for x in cmd)
        else:
            cmd_string = ' '.join(cmd)

        log(self.name + ': ' + cmd_string)
        result = self.ssh.run(cmd_string, env=env, stdin=stdin, stdout=stdout, stderr=stderr)
        if result != 0:
            if catch:
                print('error: {} failed ({})'.format(cmd[0], result), file=logstream)
            else:
                raise subprocess.CalledProcessError(result, cmd_string)

        if return_stdout:
            return stdout.getvalue().strip()

    def assert_resource_exists(self, cls, name):
        resp = requests.get('http://{}:8337/api/v2/{}'.format(self.addr, cls))
        try:
            resources = resp.json()
        except:
            raise RuntimeError('could not parse response for {} resource {}: {}'.format(cls, name, resp.text))
        id_field = id_map[cls]
        for resource in resources:
            if id_field not in resource:
                raise RuntimeError(
                    'ASSERT: got invalid response for {} resource {} ({})'.format(cls, name, resources))
            if resource[id_field] == name:
                return

        raise RuntimeError(
            'ASSERT: {} resource {} should exist, but not found (resources: {})'.format(cls, name, resources))

    def assert_resource_not_exists(self, cls, name):
        resp = requests.get('http://{}:8337/api/v2/{}'.format(self.addr, cls))
        try:
            resources = resp.json()
        except:
            raise RuntimeError('could not parse response for {} resource {}: {}'.format(cls, name, resp.text))
        id_field = id_map[cls]
        for resource in resources:
            if id_field not in resource:
                raise RuntimeError(
                    'ASSERT: got invalid response for {} resource {} ({})'.format(cls, name, resources))
            if resource[id_field] == name:
                raise RuntimeError(
                    'ASSERT: {} resource {} should NOT exist, but found (resources: {})'.format(cls, name, resources))

    def wait_for_path(self, path: str, timeout: int = 30):
        """
        Wait for a path (e.g. a device node created by udev) to appear.
        """
        self.run(['sh', '-c', 'for i in $(seq {}); do [ -e {} ] && exit 0; sleep 1; done; exit 1'.format(
            timeout, shlex.quote(path))])

    def assert_pattern(self, target: str, checksum: str, size_mb: int = 16):
        """
        Read data back from a file or block device, bypassing the page cache,
        and compare it against a checksum.
        """
        got = self.run(['sh', '-c', 'dd if={} bs=1M count={} iflag=direct status=none | sha256sum'.format(
            shlex.quote(target), size_mb)], return_stdout=True).split()[0]
        if got != checksum:
            raise RuntimeError('ASSERT: data read from {} does not match what was written ({} != {})'.format(
                target, got, checksum))
        log('Data on {} verified'.format(target))

    def nfs_mount(self, server_ip: str, export_path: str) -> str:
        """
        Mount an NFS export.

        :returns: the path of a file on the export to run I/O against
        """
        self.run(['mkdir', '-p', IO_MOUNTPOINT])
        self.run(['mount', '-t', 'nfs', '{}:{}'.format(server_ip, export_path), IO_MOUNTPOINT])
        return os.path.join(IO_MOUNTPOINT, 'gatewaytest.bin')

    def nfs_umount(self):
        self.run(['umount', IO_MOUNTPOINT])

    def iscsi_login(self, portal_ip: str, iqn: str, lun: int = 1, port: int = 3260) -> str:
        """
        Log in to an iSCSI target.

        :returns: the block device of the given LUN
        """
        portal = '{}:{}'.format(portal_ip, port)
        self.run(['iscsiadm', '-m', 'discovery', '-t', 'sendtargets', '-p', portal])
        self.run(['iscsiadm', '-m', 'node', '-T', iqn, '-p', portal, '--login'])
        device = '/dev/disk/by-path/ip-{}-iscsi-{}-lun-{}'.format(portal, iqn, lun)
        self.wait_for_path(device)
        return device

    def iscsi_logout(self, portal_ip: str, iqn: str, port: int = 3260):
        portal = '{}:{}'.format(portal_ip, port)
        self.run(['iscsiadm', '-m', 'node', '-T', iqn, '-p', portal, '--logout'])
        self.run(['iscsiadm', '-m', 'node', '-o', 'delete', '-T', iqn, '-p', portal])

    def nvme_connect(self, addr: str, nqn: str, port: int = 4420) -> str:
        """
        Connect to an NVMe-oF subsystem over TCP.

        :returns: the block device of its first namespace
        """
        self.run(['modprobe', 'nvme-tcp'])
        self.run(['nvme', 'connect', '-t', 'tcp', '-a', addr, '-s', str(port), '-n', nqn])

        find_namespace = """
for i in $(seq 30); do
    for s in /sys/class/nvme-subsystem/nvme-subsys*; do
        [ "$(cat "$s/subsysnqn" 2>/dev/null)" = {nqn} ] || continue
        for ns in "$s"/nvme[0-9]*n[0-9]*; do
            [ -e "$ns" ] && echo "/dev/${{ns##*/}}" && exit 0
        done
    done
    sleep 1
done
exit 1
""".format(nqn=shlex.quote(nqn))
        device = self.run(['sh', '-c', find_namespace], return_stdout=True)
        self.wait_for_path(device)
        return device

    def nvme_disconnect(self, nqn: str):
        self.run(['nvme', 'disconnect', '-n', nqn])

    def start_io(self, target: str, size_mb: int = 16) -> str:
        """
        Start writing a random pattern to a file or block device in a
        background loop, with direct I/O and an fsync per round. The loop counts
        completed rounds and stops, leaving a marker, as soon as a write fails.

        :returns: the sha256 checksum of the pattern
        """
        self.run(['rm', '-f', IO_COUNT, IO_FAILED])
        self.run(['dd', 'if=/dev/urandom', 'of=' + IO_PATTERN, 'bs=1M', 'count={}'.format(size_mb), 'status=none'])
        checksum = self.run(['sha256sum', IO_PATTERN], return_stdout=True).split()[0]

        loop = ('n=0; while dd if={pattern} of={target} bs=1M oflag=direct conv=fsync status=none; do '
                'n=$((n+1)); echo $n > {count}; done; touch {failed}').format(
            pattern=IO_PATTERN, target=shlex.quote(target), count=IO_COUNT, failed=IO_FAILED)
        self.io_pid = self.run(['sh', '-c', 'setsid sh -c {} </dev/null >/dev/null 2>&1 & echo $!'.format(
            shlex.quote(loop))], return_stdout=True)
        log('Started I/O on {} (pid {})'.format(target, self.io_pid))
        self.assert_io_progress()
        return checksum

    def io_rounds(self) -> int:
        out = self.run(['sh', '-c', 'cat {} 2>/dev/null || echo 0'.format(IO_COUNT)], return_stdout=True)
        return int(out or 0)

    def assert_io_progress(self, rounds: int = 2, timeout: int = 300):
        """
        Assert that the background I/O loop has not failed and completes at
        least the given number of further rounds within the timeout.
        """
        want = self.io_rounds() + rounds
        for _ in range(timeout):
            if self.run(['sh', '-c', '[ -e {} ] && echo failed || true'.format(IO_FAILED)], return_stdout=True):
                raise RuntimeError('ASSERT: I/O on {} failed'.format(self.name))
            done = self.io_rounds()
            if done >= want:
                log('I/O on {} progressing ({} rounds)'.format(self.name, done))
                return
            time.sleep(1)
        raise RuntimeError('ASSERT: I/O on {} stalled at {} rounds for {}s'.format(
            self.name, self.io_rounds(), timeout))

    def stop_io(self, target: str, checksum: str):
        """
        Stop the background I/O loop, assert that it never failed and verify
        that the target holds the pattern.
        """
        self.assert_io_progress(rounds=1)
        self.run(['sh', '-c', 'kill -TERM -- -{pid}; while kill -0 {pid} 2>/dev/null; do sleep 0.2; done'.format(
            pid=self.io_pid)])
        # Rewrite once, since the loop may have been interrupted mid-write.
        self.run(['dd', 'if=' + IO_PATTERN, 'of=' + target, 'bs=1M', 'oflag=direct', 'conv=fsync', 'status=none'])
        self.assert_pattern(target, checksum)

    def prevent_promotion(self, resource: str):
        """
        Mask the drbd-reactor target of a resource on this node until reboot,
        so that the resource is never promoted here.
        """
        self.run(['systemctl', 'mask', '--runtime', 'drbd-services@{}.target'.format(resource)])

    def reactor_snippet(self, resource: str) -> str:
        """
        :return: the path of the drbd-reactor snippet that promotes the resource
        """
        return self.run(['sh', '-c', 'grep -lE {} /etc/drbd-reactor.d/linstor-gateway-*.toml'.format(
            shlex.quote(r'resources\."?{}"?\]'.format(re.escape(resource))))], return_stdout=True)

    def drbd_ports(self, resource: str) -> set:
        """
        :return: the TCP ports DRBD uses for the resource's connections
        """
        config = self.run(['drbdsetup', 'show', resource], return_stdout=True)
        ports = set(int(p) for p in re.findall(r'_this_host\s+\S+\s+\S+:(\d+);', config))
        if not ports:
            raise RuntimeError('No DRBD ports found for resource {}:\n{}'.format(resource, config))
        return ports

    @contextmanager
    def drbd_isolated(self, resource: str):
        """
        Drop all DRBD replication traffic of a resource on this node, so that it
        loses quorum while LINSTOR and SSH keep working. The rules are removed
        again when the context is left.
        """
        rules = []
        for port in self.drbd_ports(resource):
            for chain, direction in (('INPUT', '--dport'), ('INPUT', '--sport'),
                                     ('OUTPUT', '--dport'), ('OUTPUT', '--sport')):
                rules.append([chain, '-p', 'tcp', direction, str(port), '-j', 'DROP'])

        log('Isolating DRBD resource {} on {}'.format(resource, self.name))
        for rule in rules:
            self.run(['iptables', '-I'] + rule)
        try:
            yield
        finally:
            log('Reconnecting DRBD resource {} on {}'.format(resource, self.name))
            for rule in rules:
                self.run(['iptables', '-D'] + rule)


def check_failover(ls: LinstorConnection, nodes: Nodes, resource: str, active_node: str, connect, disconnect):
    """
    Exercise a resource from a client across both kinds of failover:

    1. A planned switchover (drbd-reactorctl evict): the client stays
       connected with background I/O running, which must keep making progress.
    2. An unplanned failover, interrupting the DRBD connections of the
       primary: the client reconnects while the old primary is cut off, and
       must find the data written before and be able to write.

    The client node is kept from being promoted, so it stays remote from the
    primary throughout.

    :param connect: connect(client) connects the client and returns the file
                    or block device to run I/O against
    :param disconnect: disconnect(client) disconnects the client again
    """
    client = nodes.client_for(active_node)
    client.prevent_promotion(resource)

    target = connect(client)
    checksum = client.start_io(target)
    primary = nodes.by_name(active_node)
    # LIO only removes a LUN once connected initiators time out, which takes
    # longer than evict's default of 20s to wait for a peer to take over.
    primary.run(['drbd-reactorctl', 'evict', '--delay', '120', primary.reactor_snippet(resource)])
    new_node = ls.wait_for_failover(resource, active_node)
    log('Resource {} switched over from {} to {}'.format(resource, active_node, new_node))
    ls.wait_inuse_stable(resource, new_node)
    client.stop_io(target, checksum)
    disconnect(client)

    active_node = new_node
    with nodes.by_name(active_node).drbd_isolated(resource):
        new_node = ls.wait_for_failover(resource, active_node)
        log('Resource {} failed over from {} to {}'.format(resource, active_node, new_node))
        target = connect(client)
        client.assert_pattern(target, checksum)
        checksum = client.start_io(target)
        client.stop_io(target, checksum)
        disconnect(client)

    ls.wait_for_resource_active(resource, retries=120)
    ls.wait_inuse_stable(resource, new_node)
    log('Resource {} healthy again, in use on {}'.format(resource, new_node))


def handle_exception(callback):
    """
    Run the specified callback after calling the original excepthook.
    This is intended to be assigned to sys.excepthook.
    :param callback: a function to be called after the original excepthook
    :return: a function that can be assigned to sys.excepthook
    """

    def handle(exc_type, exc_value, exc_traceback):
        __real_excepthook(exc_type, exc_value, exc_traceback)
        callback()

    return handle


def setup():
    parser = argparse.ArgumentParser()
    parser.add_argument('--logdir')
    parser.add_argument('node', nargs='*')
    args = parser.parse_args()

    job = re.sub(r'.*/(.*?)(?:\.py)?$', r'\1', sys.argv[0]) + '-' + time.strftime('%Y%m%d-%H%M%S')
    job_symlink = None
    if args.logdir is None:
        args.logdir = os.path.join('log', job)
        job_symlink = re.sub(r'-[^-]*-[^-]*?$', '', job) + '-latest'

    # no log() here yet, only gets initialized later
    print('Logging to directory %s' % args.logdir, file=sys.stderr)

    if not os.access(args.logdir, os.R_OK + os.X_OK + os.W_OK):
        os.makedirs(args.logdir)
    if job_symlink is not None:
        try:
            os.remove(os.path.join('log', job_symlink))
        except OSError as e:
            if e.errno != errno.ENOENT:
                raise e
        os.symlink(job, os.path.join('log', job_symlink))

    logfile = open(os.path.join(args.logdir, 'test.log'), 'w', encoding='utf-8')
    # no need to close logfile - it is kept open until the program terminates
    global logstream
    logstream = Tee()
    logstream.add(sys.stderr)
    logstream.add(logfile)

    nodes = Nodes()
    for n in args.node:
        new_node = Node(n, args.logdir)
        nodes.append(new_node)
        log('New node: {} {}'.format(new_node.name, new_node.hostname))

    # make sure that the nodes are cleaned up even if an exception occurs
    sys.excepthook = handle_exception(nodes.cleanup)
    return nodes
