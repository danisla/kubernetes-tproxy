# tproxy-podwatch

Kubernetes controller to watch for annoated pods and add/remove local node firewall rules.

Based on the [workqueue controller example source](https://github.com/kubernetes/kubernetes/blob/edce96c5b6bd4cee6ae6c05934e5078b0920d143/staging/src/k8s.io/client-go/examples/workqueue/main.go).

Must be run with `pod.spec.containers.securityContext.privileged: true` and `pod.spec.hostNetwork: true`

## Development workflow

Code is built and run in-cluster. Use the `kubectl cp` command to get new source into the pod and the `kubectl exec` command to `go run` the controller in the cluster.

Deploy the dev deployment:

```
kubectl create -f dev-deployment.yaml
```

Configure the dev pod:

```
pod=$(kubectl get pod --selector=run=tproxy-podwatch -o jsonpath='{.items..metadata.name}')
kubectl exec -it ${pod} -- mkdir -p /go/src/github.com/danisla/tproxy-podwatch/
kubectl cp main.go ${pod}:/go/src/github.com/danisla/tproxy-podwatch/main.go
kubectl exec -it ${pod} -- sh -c 'cd /go/src/github.com/danisla/tproxy-podwatch/ && go get ./...'
```

Run the controller:

```
./test.sh
```

Now, make changes to main.go and re-run `./test.sh`
