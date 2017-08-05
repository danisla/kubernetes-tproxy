#!/bin/bash

set -x
set -eo pipefail

export POD_IP=$(hostname -i)
export POD_NAME=$(hostname)

# pod egress deny
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 53 -j ACCEPT
iptables -w -t filter -A OUTPUT -p udp -m udp --dport 53 -j ACCEPT
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 443 -j ACCEPT
iptables -w -t filter -A OUTPUT -p tcp -m tcp --dport 80 -j ACCEPT
iptables -w -t filter -I OUTPUT -p tcp -m tcp --dport 1080 --destination ${HOST_IP} -j ACCEPT
iptables -w -t filter -A OUTPUT -j DROP