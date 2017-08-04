# mitmproxy

- Transparent proxy
- http proxy
- https proxy
- full https intercept if root-ca included in app containers
- egress blocked using iptables

## Usage

Note that this chart requires a cluster with alpha features enabled in order to use the custom initializer.

### Install the helm chart

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

### Using with deployments

Add this metadata annotation to your deployment spec:

```yaml
metadata:
  annotations:
    "initializer.kubernetes.io/tproxy": "true"
```

This will cause the tproxy-initializer to do the following:

- Inject the tproxy sidecar container into the pod.
- Adds default environment variables `http_proxy` and `https_proxy` that point to the mitmproxy daemonset on the same node as the pod.
- Inject the CA cert bundle with the mitmproxy CA cert to the following paths:

```
# Fedora/Redhat/CentOS
/etc/pki/tls/certs/ca-bundle.crt

# Debian/Ubuntu
/etc/ssl/certs/ca-certificates.crt
```