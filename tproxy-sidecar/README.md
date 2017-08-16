# tproxy-sidecar

Sidecar run in pod to add iptables rules to block egress traffic for the pod.

Must be run with `pod.spec.containers.securityContext.privileged: true`.

If the `HOST_IP` environment variable is present, an additional iptables rule will be added to allow traffic to `HOST_IP:1080`. This is use to pass traffic to a mitmproxy instance running in standard mode listening the node host port 1080.