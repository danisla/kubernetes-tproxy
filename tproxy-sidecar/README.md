# tproxy-sidecar

Sidecar run in pod to add iptables rules to block egress traffic and make REST call to tproxy-service to add REDIRECT rule on the local node.

When the pod terminates, the sidecar makes another REST call to the tproxy-service to remove the REDIRECT rule.