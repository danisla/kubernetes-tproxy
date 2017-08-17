# Kubernetes initializer for transparent proxy

Initializes deployments with the annotation: `initializer.kubernetes.io/tproxy: true`. Resources added to a deployment are provided via a configmap. Items that can be injected are init containers, volumes, volume mounts, and environment variables.

Based on the [Kubernetes Initializer Tutorial](https://github.com/kelseyhightower/kubernetes-initializer-tutorial) from Kelsey Hightower.

## Configmap schema

The configmap containing the initialization data and the namespace for the configmap is passed via the `-configmap` and `-namespace` arguments.

The data items from the configmap used are listed below:

- `config.containers`: List of init container specs added to the deployment. (`kubectl explain pods.spec.initContainers`)
- `config.volumes`: List of volume specs added to the deployment. (`kubectl explain pods.spec.volumes`)
- `config.volumeMounts`: List of volume mount specs added to each container in the deployment pod spec. (`kubectl explain pods.spec.containers.volumeMounts`)
- `config.envVars`: List of environment variable specs added to each container in the deployment pod spec. (`kubectl explain pods.spec.containers.env`)

## Initializer actions

For deployments with matching annotations, the following actions will be taken.

1. Append `config.containers` to the pod spec of the original deployment.
2. Append `config.volumes` to the pod spec of the original deplyoment.
3. Append `config.volumeMounts` to each container spec of the original deployment.
4. Append `config.envVars` to each container spec of the original deployment.
5. Append the `initializer.kubernetes.io/tproxy: true` annotation to the pod spec of the original deployment. This is needed so that the `tproxy-podwatch` controller can trigger based on the pod after the deployment is initialized.