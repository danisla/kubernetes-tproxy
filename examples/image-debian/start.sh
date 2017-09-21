#!/bin/bash

_term() { 
    echo "Caught SIGTERM signal, exiting."
    exit 0
}

trap _term SIGTERM

# main loop
(while true; do
    echo "`date -u` https://www.google.com: `curl -s -o /dev/null -w '%{http_code}' https://www.google.com`"
    echo "`date -u` https://storage.googleapis.com/solutions-public-assets/: `curl -s -o /dev/null -w '%{http_code}' https://storage.googleapis.com/solutions-public-assets/adtech/dfp_networkimpressions.py`"
    echo "`date -u` `ping -c1 www.google.com 2>&1`"
    sleep 5
done) &

child=$! 
wait "$child"