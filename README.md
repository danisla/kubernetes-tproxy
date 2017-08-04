# Kubernetes Transparent Proxy Example

Example of how to deploy a transparent proxy to filter and intercept all http/s traffic out of a pod.

This is done using the [`tproxy-initializer`](./tproxy-initializer) Kubernetes Initializer to inject a sidecar container, configmap and environment variables into a deployment when the annotation `"initializer.kubernetes.io/tproxy": "true"` is present. 

The purpose of the [`tproxy-sidecar`](./tproxy-sidecar) container is to create iptables rules in the pod network to block egress traffic out of the pod and to tell the [`tproxy-service`](./tproxy-service) to add a REDIRECT firewall rule on the node network for the pod to the mitmproxy service. When the pod is terminated, the REDIRECT rule is removed by making a similar request to the `tproxy-service`.

Technology used:

- [Google Container Engine](https://cloud.google.com/container-engine/)
- [Kubernetes Initializers](https://kubernetes.io/docs/admin/extensible-admission-controllers/#what-are-initializers)
- [mitmproxy](https://mitmproxy.org/)

Special thanks to the [Kubernetes Initializer Tutorial](https://github.com/kelseyhightower/kubernetes-initializer-tutorial) by Kelsey Hightower for the Go example.

**Figure 1.** *tproxy diagram*

<img src="./diagram.png" width="800px"></img>

## Example

### Create alpha GKE cluster

As of GKE 1.7.2 the initializers feature is alpha and requires an alpha GKE cluster.

```
gcloud container clusters create dev --machine-type n1-standard-4 --num-nodes 3 --enable-kubernetes-alpha --cluster-version 1.7.2
```

### Build the container images

Use [Container Builder](https://cloud.google.com/container-builder/docs/) to build the container images. This will place the images in your current project.

```
cd tproxy-initializer && ./build-container && cd -

cd tproxy-service && ./build-container && cd -

cd tproxy-sidecar && ./build-container && cd -

cd example-app/image-debian && ./build-container && cd -

cd example-app/image-centos && ./build-container && cd -
```

### Deploy tproxy helm chart

Gen certs:

```
docker run -it --rm -v ${PWD}/mitmproxy/certs/:/home/mitmproxy/.mitmproxy mitmproxy/mitmproxy
```

```
helm init
```

```
helm install -n tproxy .
```

### Deploy sample apps

Run the sample apps to demonstrate using and not using the annotation to trigger the initializer. There are variants for debian and centos to show how the mitmproxy ca certs are mounted per distro.

```
kubectl create -f example-app/
```

Here is the YAML for the deployment with the annotation:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: debian-app-locked
  annotations:
    "initializer.kubernetes.io/tproxy": "true"
spec:
  replicas: 1
  selector:
    matchLabels:
      run: app
  template:
    metadata:
      labels:
        run: app
        variant: debian-locked
    spec:
      containers:
        - name: app
          image: gcr.io/disla-goog-com-csa-ext/example-app:debian
          imagePullPolicy: Always
```

### Test output

Pod without tproxy:

```
kubectl logs $(kubectl get pods --selector=variant=debian -o=jsonpath={.items..metadata.name}) -c app
```

> Traffic to https endpoints is unrestricted.

Pod with tproxy:

```
kubectl logs $(kubectl get pods --selector=variant=debian-locked -o=jsonpath={.items..metadata.name}) -c app
```

> All http/s traffic is proxied through mitmproxy, only the route to the google storage bucket is permitted per the mitmproxy python script. All other egress traffic is blocked.


## Cleanup

Delete the sample apps:

```
kubectl delete -f exmaple-apps/
```

Delete the tproxy helm release:

```
helm delete --purge tproxy
```

Delete the GKE cluster:

```
gcloud container clusters delete dev
```