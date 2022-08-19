# Setting up the cluster

> Container runtime has been setup using [this](https://gist.github.com/xvzf/11fca11491188c20a39afa803a2f3240) cloud-init.

> Requires 7 VMs: 192.168.88.{41,42,43,44,45,46,47}; one master, 5 workers; 1 benchmark instance. Be sure swap is disabled; tested with Kubernetes 1.24

> The monitoring stack is configured to remote-write to `oasis.xvzf.tech/prometheus`; this likely has to be adapted. A long-term retention installation of [Grafana Mimir](https://github.com/grafana/mimir) is recommended

Steps:
1. `kubeadm init --pod-network-cidr=10.33.0.0/16 --service-cidr=10.96.0.0/16`
2. Join nodes with `<command retrieved from step 1>`, e.g.: `kubeadm join 192.168.88.41:6443 --token y1h2c3.hd97hyr0jhp2owu0 --discovery-token-ca-cert-hash sha256:bb4f07e200e47ba246932f6f1d93b168a926673611991bfce5b026f6fa8d793a`
3. Taint the benchmark node so it stays free of workloads: `k taint node k8s-bench0 fortio-only=true:NoSchedule`
4. Setup networking (assumes L2 connectivity): `kubectl apply -f flannel.yaml`
5. Setup Cert-Manager: `kubectl apply -f /overlays/target/apps/cert-manager/`
6. Setup local-path storage provider `kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.22/deploy/local-path-storage.yaml` and mark it as default storage class
7. Setup Monitoring: `kubectl apply -f /base/apps/grafana-agent/`
8. Deploy lightpath: `kubectl apply -k <repo_root>/deploy/default
