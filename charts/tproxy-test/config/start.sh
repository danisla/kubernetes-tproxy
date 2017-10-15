#!/bin/bash

function checkURLStatusCode() {
    STATUS=$1
    URL=$2
    ARGS=$3
    RES=$(curl -s -o /dev/null -w '%{http_code}' ${ARGS} "${URL}")
    [[ ${RES} == ${STATUS} ]]
}

function testRes() {
    RES=$1
    MSG=$2

    T=PASS; [[ $RES -ne 0 ]] && T=FAIL
    echo "${T}:  ${MSG}"

    [[ "${T}" != "PASS" ]]
}

PASSED=true

# Allowed egress https
checkURLStatusCode 200 https://storage.googleapis.com/solutions-public-assets/adtech/dfp_networkimpressions.py
testRes $? "Egress to allowed https should return status 200 (https://storage.googleapis.com/solutions-public-assets/)." || PASSED=false

# Blocked egress
checkURLStatusCode 418 https://www.google.com
testRes $? "Egress to external https should return status 418 (https://www.google.com)." || PASSED=false

# GCE Metadata server
checkURLStatusCode 418 http://metadata.google.internal
testRes $? "GCE metadata server should return status 418 (http://metadata.google.internal)." || PASSED=false

# K8S cluster service
nc -vz -w1 kubernetes-dashboard.kube-system.svc.cluster.local 80 2>&1 | grep -q "timed out"
testRes $? "K8S cluster service should time out (kubernetes-dashboard.kube-system.svc.cluster.local)." || PASSED=false

# K8S API token test
test ! -e /var/run/secrets/kubernetes.io/serviceaccount/token
testRes $? "K8S API token from service account should not exist." || PASSED=false

# Ping test
ping -c1 www.google.com 2>&1 | grep -q "Operation not permitted"
testRes $? "ICMP (ping) operation should not be permitted." || PASSED=false

if [[ $PASSED ]]; then
    echo "INFO:  All tests passed"
else
    echo "ERROR: At least one test failed."
fi

[[ $PASSED ]]