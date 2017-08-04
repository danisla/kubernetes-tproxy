#!/bin/bash

_term() { 
    echo "Caught SIGTERM signal, exiting."
    exit 0
}

trap _term SIGTERM

# main loop
(while true; do
    echo "`date -u` External IP: `curl -s https://api.ipify.org`"
    echo "`date -u` GCS Bucket object content length: `curl -sI https://storage.googleapis.com/solutions-public-assets/adtech/dfp_networkimpressions.py |grep -i x-goog-stored-content-length`"
    sleep 5
done) &

child=$! 
wait "$child"