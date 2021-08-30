#! /usr/bin/env python

import argparse
import subprocess
import sys

class Nodes(list):
    def __init__(self, members=[]):
        super(Nodes, self).__init__(members)

    def run(self, command):
        return [ n.run(command) for n in self ]

class Node:
    def __init__(self, host):
        ip = self._run_scalar(host, ['hostname', '-i'])
        hostname = self._run_scalar(host, ['hostname'])
        self.ip = ip
        self.hostname = hostname

    def _run_scalar(self, host, command):
        lines = self._run(host, command)
        stdout = lines[0]
        stderr = lines[1]
        if len(stderr) > 0:
            raise Exception('\n'.join([ 'ssh error: {}'.format(l.strip().decode('utf-8')) for l in stderr ]))
        return stdout[0].strip().decode('utf-8')

    def _run(self, host, command):
        ssh = subprocess.Popen(['ssh', 'root@{}'.format(host), *command],
            shell=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE)
        return ssh.stdout.readlines(), ssh.stderr.readlines()

    def run(self, command):
        return self._run(self.ip, command)


def setup():
    parser = argparse.ArgumentParser()
    parser.add_argument('node', nargs='*')
    args = parser.parse_args()

    nodes = Nodes()
    for n in args.node:
        new_node = Node(n)
        nodes.append(new_node)
        print('New node: {} {}'.format(new_node.ip, new_node.hostname))
    return nodes
