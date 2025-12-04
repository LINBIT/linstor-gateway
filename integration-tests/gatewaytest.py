#! /usr/bin/env python3

import argparse
import errno
import ipaddress
import os
import pipes
import re
import socket
import subprocess
import sys
import time
from io import StringIO
from threading import Thread

import linstor
import requests
from lbpytest.controlmaster import SSH

# stream to write output to
logstream = None

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
            cmd_string = ' '.join(pipes.quote(str(x)) for x in cmd)
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
