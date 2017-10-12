# Kubernetes isolation test

Runs a job on each of the nodes to verify tproxy is providing proper isolation.

## Running

1. Install helm with RBAC support:

        kubectl create serviceaccount tiller --namespace kube-system
        kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
        helm init --service-account=tiller

2. Create namespace for test workload:

        kubectl create namespace tproxy-test

3. Install tproxy chart:
        
        helm install --namespace tproxy-test -n tproxy --set tproxy.useRBAC=true ../tproxy

4. Get cluster info:

        CLUSTER=dev
        SVC_CIDR=$(gcloud container clusters describe $CLUSTER --zone us-central1-c --format='value(servicesIpv4Cidr)')
        DNS_IP=$(kubectl get svc kube-dns -n kube-system -o jsonpath="{.spec.clusterIP}")
        NUM_NODES=$(gcloud container clusters describe $CLUSTER --zone us-central1-c --format='value(currentNodeCount)')

5. Install tproxy-test chart:
        
        helm install --namespace tproxy-test --name tproxy-test --set blockSvcCIDR=${SVC_CIDR},allowDNS=${DNS_IP},numNodes=${NUM_NODES} .

6. Get the logs for the test job:

        kubectl logs --namespace tproxy-test job/tproxy-test-tproxy-test

    Expected output:

        PASS:  Egress to allowed https should return status 200 (https://storage.googleapis.com/solutions-public-assets/)
        PASS:  Egress to external https should return status 418 (https://www.google.com).
        PASS:  GCE metadata server should return status 418 (http://metadata.google.internal).
        PASS:  K8S Cluster service should time out (kubernetes-dashboard.kube-system.svc.cluster.local).
        PASS:  K8S API token from service account should not exist.
        PASS:  Ping operation should not be permitted.
        INFO:  All tests passed

7. Delete the tprox-test release:

        helm delete --purge tproxy-test

8. Delete the namespace:

        kubectl delete namespace tproxy-test

9. Delete the tproxy release:

        helm delete --purge tproxy