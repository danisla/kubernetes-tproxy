#!/bin/bash

export POD_IP=$(hostname -i)
export POD_NAME=$(hostname)

_term() { 
    echo "Caught SIGTERM signal! Cleaning up firewall"
    iptables -t filter -I OUTPUT -p tcp -m tcp --dport 9000 --destination ${HOST_IP} -j ACCEPT
    curl -sf http://${HOST_IP}:9000/tproxy -d action=remove -d pod_ip=${POD_IP} -d pod_name=${POD_NAME}
    [[ $? -ne 0 ]] && echo "Error cleaning up firewall" && exit 1
    sleep 2
    exit 0
}

trap _term SIGTERM

echo "Adding transparent proxy firewall for pod ${POD_NAME}, ${POD_IP}"
curl -sf http://${HOST_IP}:9000/tproxy -d action=add -d pod_ip=${POD_IP} -d pod_name=${POD_NAME}
[[ $? -ne 0 ]] && echo "Error creating firewall" && exit 1

# pod egress deny
iptables -t filter -A OUTPUT -p tcp -m tcp --dport 53 -j ACCEPT
iptables -t filter -A OUTPUT -p udp -m udp --dport 53 -j ACCEPT
iptables -t filter -A OUTPUT -p tcp -m tcp --dport 443 -j ACCEPT
iptables -t filter -A OUTPUT -p tcp -m tcp --dport 80 -j ACCEPT
iptables -t filter -I OUTPUT -p tcp -m tcp --dport 1080 --destination ${HOST_IP} -j ACCEPT
iptables -t filter -A OUTPUT -j DROP

echo "Waiting for pod termination"
(while true; do sleep 1000; done) &
child=$! 
wait "$child"