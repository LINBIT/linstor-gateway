#!/bin/bash

ssh-keygen -f /root/.ssh/id_rsa -y >/root/.ssh/id_rsa.pub

sorted_targets=$(echo "$TARGETS" | tr ',' '\n' | sort)
nodes=()
for target in $sorted_targets; do
	host=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $target hostname)
	nodes+=("$host")
done

tests/${TEST_NAME}.py --logdir /log ${nodes[@]}
