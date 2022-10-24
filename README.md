# Lightpath - a kube-proxy replacement based on Envoy

Lightpath consists of multiple data plane and control plane components:
1. **Control Plane**:
  - `controlplane` (node-scoped): runs on every node and acts as control server for Envoy
  - `webook` (cluster-scoped): The webhook components intercepts the service creation and updates the handling proxy to Lightpath. This allows a hybrid deployment next to the kube-proxy
2. **Data Plane**:
  - `proxy`(node-scoped): Envoy acts as proxy instance and is configured to connect to the node-local control plane
  - `redirect` configures the iptables redirect targets for ClusterIPs handled by Lightpath

# Limitations
As a PoC, the TCP and HTTP protocol are targeted. STCP and UDP based protocols are not supported (yet) and need to be handled by kube-proxy. Therefore, a hybrid setup is required for full service connectivity.

# Deployment

Lightpath can be deployed on any Kubernetes Cluster with an iptables based CNI (e.g. Calico, Flannel, ...).

## Test cluster
The `make k8s-up` command bootstraps a new [Kubernetes in Docker](https://kind.sigs.k8s.io/docs/user/quick-start/), which can be used for further testing

## Deploying
> Lightpath images are (publicly) available for linux/amd64 and linux/arm64.

The deployment manifests are based on [Kustomize](http://kustomize.io) and are generated with `kustomize build deploy/default/`. Lightpath requires cert-manager to be installed on the tareted cluster; the installation will fail otherwise. The `make k8s-up` command already setups all required dependencies.

As a shortcut, `make deploy` applies the kustomize manifests to Kubernetes results.

A test deployment running a simple echo server can be applied to a cluster with `kubectl apply -f hack/ci/test-deploy.yaml`.
A new pod with an interactive shell can be bootstrapped using e.g. `kubectl run tmp-net-debug-shell -it --rm --image nicolaka/netshoot -- /bin/bash`, which can then be used to access the deployed service:
```
bash-5.1# curl whoami.whoami.svc
Hostname: whoami-5978d4b87d-dwx2t
IP: 127.0.0.1
IP: ::1
IP: 10.244.2.3
IP: fe80::e8f1:57ff:fefa:2e66
RemoteAddr: 172.18.0.3:42850
GET / HTTP/1.1
Host: whoami.whoami.svc
User-Agent: curl/7.83.1
Accept: */*
X-Envoy-Attempt-Count: 1
X-Envoy-Expected-Rq-Timeout-Ms: 5000
X-Forwarded-Proto: http
X-Request-Id: c686bc0b-1e33-445c-b945-9ae883313303
```

The X-Envoy headers indicate that the request has been successfully handled by lightpath. Furthermore, the Envoy admin interface can ben accessed by e.g. port-forwarding the configured admin port (15000) to the local machine:
```bash
$ kubectl port-forward $(kubectl get pod -oname -l app.kubernetes.io/name=proxy | head -n 1) 15000:15000 &

$ # E.g. retrieve the configuration dump:
$ curl http://localhost:15000/config_dump
# ...
```

# Repository Structure

Overall, the repository acts as mono-repository for all lightpath components, configuration and performed benchmarks performed throughout the master thesis.

The actual application source code is distributed in the `cmd`, `pkg` and `internal` folders, with the `cmd/{controlplane,redirect}` holding the two Golang based services.
The webhook is implemented in the Rego language as part of the [open policy agent](https://www.openpolicyagent.org) in the Kustomize based Kubernetes manifests (`deploy/{proxy,redirect,webhook,controlplane}`). The `deploy/default` Kustomization acts as meta package for all components.

The main Golang packages are located in `pkg/state/` and `pkg/translations/`. The translation package defines the actual mapping between Kubernetes resources and the Envoy representation.

Performed benchmark/load tests and scalability analysis is included in the `hack/benchmark` folder, defining the infrastructure in `hack/benchmark/bench-infra` distributed across a monitoring cluster and workload cluster. The actual experiments performed are located alongside the [Fortio](https://fortio.org) results in `hack/benchmark/experiments/{scenarios,results}`.
All experiments base on the `hack/benchmark/experiments/generator` environment, which has been used to generate the scenarios.

The `bpf` directory includes early steps towards eBPF based forwarding of traffic to the node-local proxy and has not been covered in the thesis and is in a non functional state, but shipped for completeness.
