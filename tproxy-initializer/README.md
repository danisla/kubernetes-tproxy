# Kubernetes initializer for transparent proxy

Injects sidecar container to deployments with the annotation: `initializer.kubernetes.io/tproxy: true`

The sidecar does the following:

- Installs iptables rules for the pod network to block all egress traffic except to DNS and the mitmproxy endpoint.
- 