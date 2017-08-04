# tproxy-service

REST service that adds iptable rules to the local node.

Must be run with `pod.spec.hostNetwork: true` and `pod.spec.containers.securityContext.privileged: true`