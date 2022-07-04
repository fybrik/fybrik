# Enable Control Plane Security

<!-- TODO: once the Helm chart is ready change the text in the Mutual TLS section  -->

Kubernetes  [`NetworkPolicies`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) and optionally [Istio](https://istio.io/) are used to protect components of the control plane. Specifically, traffic to connectors that run as part of the control plane must be secured. Follow this page to enable control plane security.

## Ingress traffic policy

The installation of Fybrik applies a Kubernetes [`NetworkPolicy`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) resource to the `fybrik-system` namespace. This resource ensures that ingress traffic to connectors is only allowed from workloads that run in the `fybrik-system` namespace and thus disallow access to connectors from other namespaces or external parties.

The `NetworkPolicy` is always created. However, your Kubernetes cluster must have a [Network Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/) with `NetworkPolicy` support. Otherwise, `NetworkPolicy` resources will have no affect. While most Kubernetes distributions include a network plugin that enfoces network policies, some like [Kind](https://kind.sigs.k8s.io/) do not and require you to install a separate network plugin instead.

## Mutual TLS

### Using Fybrik

Fybrik can be configured to have the traffic between the manager and the connectors encrypted with mutual tls authentication.

To enable it, a set of helm chart fields should be set upon Fybrik deployment:
The fields contain information about the security level to use (non, TLS, or mutual TLS), and the Kubernetes secrets that contain the certificates of the manager (aka the client) and the servers (aka the data catalog and policy manager) as well as the certificates of the CAs which were used to sign the client/servers certificates.

More information about the TLS-related fields is found in Fybrik helm chart [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/values.yaml) file.

### Using Istio

Alternatively, if Istio is installed in the cluster then you can use [automatic mutual TLS](https://istio.io/latest/docs/tasks/security/authentication/authn-policy/#auto-mutual-tls) to encrypt the traffic to the connectors.

Follow these steps to enable mutual TLS:

- Ensure that Istio 1.6 or above is installed.

- Enable Istio sidecar injection in the `fybrik-system` namespace:

    ```bash
    kubectl label namespace fybrik-system istio-injection=enabled
    ```
- Create Istio `PeerAuthentication` resource to enable mutual TLS between containers with Istio sidecars:
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: "security.istio.io/v1beta1"
    kind: "PeerAuthentication"
    metadata:
    name: "premissive-mtls-in-control-plane"
    namespace: fybrik-system
    spec:
      mtls:
        mode: PERMISSIVE    
    EOF
    ```
- Create Istio `Sidecar` resource to allow any egress traffic from the control plane containers:
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Sidecar
    metadata:
    name: sidecar-default
    namespace: fybrik-system
    spec:
    egress:
    - hosts:
        - "*/*"
    outboundTrafficPolicy:
        mode: ALLOW_ANY
    EOF
    ```
- Restart the control plane pods:
    ```bash
    kubectl delete pod --all -n fybrik-system
    ```
