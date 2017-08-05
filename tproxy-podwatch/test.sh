#!/bin/bash

# run this first:
#   kubectl create -f dev-deployment.yaml

pod=$(kubectl get pod --selector=run=tproxy-podwatch -o jsonpath='{.items..metadata.name}')
kubectl cp main.go ${pod}:/go/src/github.com/danisla/tproxy-podwatch/main.go

kubectl exec -it ${pod} -- go run /go/src/github.com/danisla/tproxy-podwatch/main.go