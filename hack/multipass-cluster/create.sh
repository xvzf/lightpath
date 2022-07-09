#!/bin/bash

MEMORY="2G"
CPU="2"
DISK="20G"
PREFIX="k8s"

while getopts "n:m:c:p:" opt; do
    case $opt in
        n) node_count="$OPTARG";;
        m) MEMORY="$OPTARG";;
        c) CPU="$OPTARG";;
        p) PREFIX="$OPTARG";;
        d) DISK="$OPTARG";;
    esac
done

if [ -z $node_count ]; then
  echo "Usage: ./$0 -n <node-count> [-c <num cpu>] [-m <memory size>] [-d <disk size>] [-p <multipass-vm-prefix>]"
fi

# configure multipass bridged network
# multipass set local.bridged-network=ens18

# create master
master_name="${PREFIX}-master"

echo "[ ] Creating master ${master_name}"
multipass launch -n "$master_name" --cloud-init=./cloud-init.yaml -c "$CPU" -m "$MEMORY" --disk "$DISK"

# Setup controlplane
multipass exec "$master_name" -- sudo kubeadm init \
    --pod-network-cidr=10.244.0.0/16 \
    --service-cidr=172.30.0.0/16 \
    --ignore-preflight-errors=NumCPU

# FIXME allow dual stack networking
# --pod-network-cidr=172.20.0.0/16,fd00:8888:1::/56 \
# --service-cidr=172.30.0.0/16,fd00:8888:2::/108 \

# Unmask master for scheduling pods
multipass exec "$master_name" -- sudo kubectl --kubeconfig /etc/kubernetes/admin.conf taint nodes --all node-role.kubernetes.io/master-

# Install CNI
multipass exec "$master_name" -- sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml

# Retrieve join command for future nodes
join_command=$(multipass exec "$master_name" -- sudo kubeadm token create --print-join-command)

echo "[+] Created master ${master_name}"

for ((i=0; i<(node_count - 1); i++)); do
    node_name="${PREFIX}-node${i}"
    echo "[ ] Creating node ${node_name}"
    # Launch worker node
    multipass launch -n "$node_name" --cloud-init=./cloud-init.yaml -c "$CPU" -m "$MEMORY" --disk "$DISK"
    # Join worker node to the cluster
    multipass exec "$node_name" -- sudo ${join_command}
    echo "[+] Created node ${node_name}"
done

# Install cert-manager
multipass exec "$master_name" -- sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml

# Write kubeconfig
echo "[+] Exporting kubeconfig"
kubeconfig=$(mktemp)
$(multipass exec "$master_name" -- sudo cat /etc/kubernetes/admin.conf) > "${kubeconfig}"

echo "[+] Cluster created, run: export KUBECONFIG=${kubeconfig}"
