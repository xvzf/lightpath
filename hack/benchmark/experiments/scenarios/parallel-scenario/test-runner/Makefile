# We're storing the k6 results in Prometheus for easy analysis
PROMETHEUS_ENDPOINT:="https://oasis.xvzf.tech/prometheus/api/v1/write"
PROMETHEUS_USER:="lightpath"
PROMETHEUS_PASSWORD:="lightpath"

LIGHTPATH_CLUSTERIP:="10.96.10.189"
KUBE_PROXY_CLUSTERIP:="10.96.143.78"

run-%: k6
        K6_PROMETHEUS_FLUSH_PERIOD=1s \
        K6_PROMETHEUS_REMOTE_URL=$(PROMETHEUS_ENDPOINT) \
        K6_PROMETHEUS_USER=$(PROMETHEUS_USER) K6_PROMETHEUS_PASSWORD=$(PROMETHEUS_PASSWORD) \
        ./k6 run script.js -o output-prometheus-remote \
        -e KUBE_PROXY_CLUSTERIP=$(KUBE_PROXY_CLUSTERIP) \
        -e LIGHTPATH_CLUSTERIP=$(KUBE_PROXY_CLUSTERIP) \
        --tag testid=$*


xk6-install:
        go install go.k6.io/xk6/cmd/xk6@latest

k6: xk6-install
        xk6 build --with github.com/grafana/xk6-output-prometheus-remote

