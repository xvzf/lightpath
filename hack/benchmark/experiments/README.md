# Load test experiments

The load test was performed with [fortio](https://github.com/fortio/fortio):
```
fortio load -qps 1000 -c 8 -jitter -t 2m -labels "<labels>" -allow-initial-errors -log-errors=false -json <result-path> http://<cluster-ip>:8080/
```
