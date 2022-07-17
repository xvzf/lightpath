# Benchmark infrastructure for lightpath

> This repository follows the recommended FluxCD structure without being reconciled by flux.


## result-storage

Prometheus based TSDB where benchmark and system metrics are stored and can be queried along. Exposed at [oasis.xvzf.tech](https://oasis.xvzf.tech) (credentials: `lightpath:lightpath`).


## bench

Benchmark cluster targeting the _target_ cluster; k6 benchmarks are run here and push results to the _result-storage_ cluster


## target

This clusters runs a dummy load-generation application with configureable topologies alongside Lightpath!
