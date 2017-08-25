#!/bin/bash

set -x
set -eo pipefail

export POD_IP=$(hostname -i)
export POD_NAME=$(hostname)

# dns allow
[[ -n "${ALLOW_DNS}" ]] && iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 53 --destination ${ALLOW_DNS} -j ACCEPT
[[ -n "${ALLOW_DNS}" ]] && iptables -w -t filter -A OUTPUT -p udp -m udp --dport 53 --destination ${ALLOW_DNS} -j ACCEPT

# service network deny
[[ -n "${BLOCK_SVC_CIDR}" ]] && iptables -w -t filter -A OUTPUT --destination ${BLOCK_SVC_CIDR} -j REJECT

# pod egress deny
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 53 -j ACCEPT
iptables -w -t filter -A OUTPUT -p udp -m udp --dport 53 -j ACCEPT
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 443 -j ACCEPT
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 80 -j ACCEPT
[[ -n "${HOST_IP}" ]] && iptables -w -t filter -I OUTPUT -p tcp -m tcp --dport 1080 --destination ${HOST_IP} -j ACCEPT
iptables -w -t filter -A OUTPUT -j REJECT