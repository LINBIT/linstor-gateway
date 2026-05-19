#!/bin/sh
# Wrapper run inside the awx-ee container by virter's provision steps.
# awx-ee ships a minimal set of collections and does not include
# community.general (which provides lvg, lvol, ...), so install it here
# before running the playbook.
set -eu

ansible-galaxy collection install -r /virter/workspace/virter/requirements.yml

exec ansible-playbook "$@"
