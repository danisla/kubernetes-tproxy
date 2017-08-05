# tproxy-sidecar

Sidecar run in pod to add iptables rules to block egress traffic for the pod.

Must be run with `pod.spec.containers.securityContext.privileged: true`.
