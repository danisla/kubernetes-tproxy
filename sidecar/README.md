# tproxy-sidecar

Sidecar run in pod to add iptables rules to block egress traffic for the pod.

Must be run with `pod.spec.containers.securityContext.privileged: true`.

If the `NODE_NAME` environment variable is present, an additional iptables rule will be added to allow traffic to `NODE_NAME:1080`. This is use to pass traffic to a mitmproxy instance running in standard mode listening the node host port 1080.