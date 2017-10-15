# Kubernetes Transparent Proxy

Transparent proxy and filtering for Kubernetes pods.

This project provides transparent proxy to pods using two deployment scenarios:

1. On any K8S cluster with manual addition of the init container.
2. A K8S 1.7+ cluster with deployment annotations and initializers to inject the init container.

The init container is responsible for adding the firewall rules to redirect outbound http/s traffic to the proxy server.

See the Helm chart [README.md](./charts/tproxy/README.md) for all chart configuration options.

Technology used:

- [Kubernetes Initializers](https://kubernetes.io/docs/admin/extensible-admission-controllers/#what-are-initializers)
- [Kubernetes Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md)
- [Kubernetes RBAC](https://kubernetes.io/docs/admin/authorization/rbac/)
- [mitmproxy](https://mitmproxy.org/)
- [Kubernetes Helm](https://github.com/kubernetes/helm)
- [Google Container Engine](https://cloud.google.com/container-engine/)

## Deploying without initializers

Kubernetes Initializers are in alpha as of 1.7. This section shows how to deploy and use the transparent proxy on a K8S 1.6 cluster.

**Figure 1.** *tproxy diagram*

<img src="./tproxy_diagram.png" width="800px"></img>

1. Install the helm chart:

```sh
cd charts/tproxy
helm install -n tproxy .
cd -
```

2. Run the example app:

```sh
kubectl apply -f examples/debian-locked-manual.yaml
```

3. Inspect the logs:

```sh
kubectl logs --selector=app=debian-app,variant=locked --tail=4
```

Example output:

```
https://www.google.com: 418
https://storage.googleapis.com/solutions-public-assets/: 200
PING www.google.com (209.85.200.147): 56 data bytes
ping: sending packet: Operation not permitted
```

## Deploying with Initializers

Using the Kubernetes Initializer simplifies the runtime configuration. The initializer automatically intercepts deployments with the annotation: "initializer.kubernetes.io/tproxy": "true"` and adds the init container to the deployment.

**Figure 1.** *tproxy with initializers diagram*

<img src="./tproxy_initializers_diagram.png" width="800px"></img>

1. Create an alpha GKE cluster with initializer support:

```sh
gcloud container clusters create tproxy-example \
  --zone us-central1-f \
  --machine-type n1-standard-1 \
  --num-nodes 3 \
  --enable-kubernetes-alpha \
  --cluster-version 1.7.6
```

> NOTE: Run `gcloud container get-server-config --zone us-central1-f` to see all cluster versions.

2. Install Helm:

```sh
curl -sL https://storage.googleapis.com/kubernetes-helm/helm-v2.5.1-linux-amd64.tar.gz | tar -zxvf - && sudo mv linux-amd64/helm /usr/local/bin/ && rm -Rf linux-amd64

helm init
```

3. Install the Helm Chart:

```sh
cd charts/tproxy
helm install -n tproxy --set tproxy.useInitializer=true .
cd -
```

4. Deploy the example app that uses the annotation:

```sh
kubectl create -f examples/debian-locked.yaml
```

5. Inspect the logs:

```sh
kubectl logs --selector=app=debian-app,variant=locked --tail=4
```

Example output:

```
https://www.google.com: 418
https://storage.googleapis.com/solutions-public-assets/: 200
PING www.google.com (209.85.200.147): 56 data bytes
ping: sending packet: Operation not permitted
```
