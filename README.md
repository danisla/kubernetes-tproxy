# Kubernetes Transparent Proxy Example

Example of how to deploy a transparent proxy to filter and intercept all http/s traffic out of a pod.

This is done using the [`tproxy-initializer`](./tproxy-initializer) Kubernetes Initializer to inject a sidecar init container, configmap and environment variables into a deployment when the annotation `"initializer.kubernetes.io/tproxy": "true"` is present. 

The purpose of the [`tproxy-sidecar`](./tproxy-sidecar) container is to create iptables rules in the pod network to block egress traffic out of the pod. The [`tproxy-podwatch`](./tproxy-podwatch) controller watches for pod changes containing the annotation and automatically add/removes the local firewall REDIRECT rules to apply the transparent proxy to the pod.

Technology used:

- [Kubernetes Initializers](https://kubernetes.io/docs/admin/extensible-admission-controllers/#what-are-initializers)
- [Kubernetes Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md)
- [Kubernetes RBAC](https://kubernetes.io/docs/admin/authorization/rbac/)
- [mitmproxy](https://mitmproxy.org/)
- [Kubernetes Helm](https://github.com/kubernetes/helm)
- [Google Container Engine](https://cloud.google.com/container-engine/)

Special thanks to the [Kubernetes Initializer Tutorial](https://github.com/kelseyhightower/kubernetes-initializer-tutorial) by Kelsey Hightower for the Go example.

**Figure 1.** *tproxy diagram*

![diagram](./diagram.png)

## Example

### Create alpha GKE cluster

As of K8S 1.7 the initializers feature is alpha and requires an alpha GKE cluster.

Create cluster with latest Kubernetes version:

```
VERSION=$(gcloud container get-server-config --format='get(validMasterVersions[0])')

gcloud container clusters create dev \
  --machine-type n1-standard-4 \
  --num-nodes 3 \
  --enable-kubernetes-alpha \
  --cluster-version $VERSION \
  --no-enable-legacy-authorization
```

### Build the container images

Use [Container Builder](https://cloud.google.com/container-builder/docs/) to build the container images. This will place the images in your current project.

```
cd tproxy-initializer && ./build-container && cd -

cd tproxy-podwatch && ./build-container && cd -

cd tproxy-sidecar && ./build-container && cd -
```

### Deploy tproxy Helm chart

Gen certs:

```
docker run -it --rm -v ${PWD}/mitmproxy/certs/:/home/mitmproxy/.mitmproxy mitmproxy/mitmproxy
```

Create service account for Helm

```
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
```

Initialize Helm:

```
helm init --service-account=tiller
```

Install the chart:

```
PROJECT_ID=$(gcloud config get-value project)

helm install -n tproxy --set images.tproxy_registry=${PROJECT_ID} mitmproxy
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
          image: danisla/example-app:debian
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

## Example Without Initializer

If you do not want to use the alpha initializer feature, you can still achieve the same sidecar behavior by adding the init container to the pod spec like the example below. You do still need a 1.7.x cluster because the sidecar uses the Downward API to reflect the node host IP into the environment and the [status.hostIP field is new to 1.7](https://github.com/kubernetes/kubernetes/issues/24657).

Make sure to pass the `--set tproxy.useInitializer=false` arg to the `helm install` command to skip installation of the initializer.

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: debian-app-locked
spec:
  replicas: 1
  selector:
    matchLabels:
      run: app
  template:
    metadata:
      annotations:
        # The podspec annotation is still needed by the podwatch controller.
        "initializer.kubernetes.io/tproxy": "true"
      labels:
        run: app
        variant: debian-locked
    spec:
      initContainers:
        - name: tproxy
          image: danisla/tproxy-sidecar:0.0.1
          imagePullPolicy: IfNotPresent
          securityContext:
            privileged: true
          env:
            - name: HOST_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.hostIP
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 100m
              memory: 64Mi
      containers:
        - name: app
          image: danisla/example-app:debian
          imagePullPolicy: Always
          volumeMounts:
            - name: ca-certs-debian
              mountPath: /etc/ssl/certs/
            - name: ca-certs-debian
              # Adding the cert to the /extra dir preserves it if update-ca-certificates is run after init.
              mountPath: /usr/local/share/ca-certificates/extra/
      volumes:
        - name: ca-certs-debian
          configMap:
              name: root-certs
              items:
                - key: root-certs.crt
                  path: ca-certificates.crt
```

The above spec also reflects a complete view of the deployment after it is modified by the initializer.

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
