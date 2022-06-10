# Limitations

- Single IP address per endpoint (no pods with multiple IPs supported)
- as a PoC, the TCP protocol is targeted

# Deployment

lightpath has two components:
- a mutating webhook, which injects the `service.kubernetes.io/service-proxy-name=lightpath` label (when not existing), creating an "opt-out" enabling of the lightpath proxy
- a proxy component running on each node, managing IPTable redirects for its services (will potentially be replaced by eBPF)
