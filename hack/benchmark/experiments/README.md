# Load test experiments

The load test was performed with [fortio](https://github.com/fortio/fortio):
```
fortio load -qps 1000 -nocatchup  -uniform -c 100 -t 5m -labels "simulated-error" -json ~/res-simulated-error.json http://10.96.36.223:8080/
```
