# Enable Control Plane Security

<!-- TODO: once the Helm chart is ready change the text in the Mutual TLS section  -->

Kubernetes  [`NetworkPolicies`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) and optionally [Istio](https://istio.io/) are used to protect components of the control plane. Specifically, traffic to connectors that run as part of the control plane must be secured. Follow this page to enable control plane security.

## Ingress traffic policy

The installation of Fybrik applies a Kubernetes [`NetworkPolicy`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) resource to the `fybrik-system` namespace. This resource ensures that ingress traffic to connectors is only allowed from workloads that run in the `fybrik-system` namespace and thus disallow access to connectors from other namespaces or external parties.

The `NetworkPolicy` is always created. However, your Kubernetes cluster must have a [Network Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/) with `NetworkPolicy` support. Otherwise, `NetworkPolicy` resources will have no affect. While most Kubernetes distributions include a network plugin that enfoces network policies, some like [Kind](https://kind.sigs.k8s.io/) do not and require you to install a separate network plugin instead.

## Mutual TLS

### Using Fybrik

Fybrik can be configured to have the traffic between the manager and the connectors encrypted by using TLS. In addition, mutual TLS authentication is possible too.
 
In the TLS mode, the connectors (aka the servers) should have their certificates available to provide them to the manager (aka client) in the TLS protocol handshake process. In mutual TLS mode, both the manager and connector should have their certificates available.


#### Adding TLS Secrets

The manager/connectors certificates are kept in Kubernetes secret:

For each component copy its certificate into a file names tls.crt. Copy the certificate key into a file named tls.key.

Use kubectl with the tls secret type to create the secrets.

```bash
kubectl -n fybrik-system create secret tls tls-manager-certs \
  --cert=tls.crt \
  --key=tls.key
```

[Cert-manager](https://cert-manager.io/) can also be used to generate the secret above using its [`Certificate`](https://cert-manager.io/docs/concepts/certificate/) resource. For example, the following is an example of a `Certificate` resource for the manager where a tls type secret named `tls-manager-certs` is created by the cert-manager. The `issuerRef` field points to a cert-manager resource names [`Issuer`](https://cert-manager.io/docs/configuration/ca/) that holds the information about the CA that signs the certificate.

```bash
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: tls-manager-cert
  namespace: fybrik-system
spec:
  dnsNames:
  - manager
  issuerRef:
    kind: Issuer
    name: ca-issuer
  secretName: tls-manager-certs
```

#### Using a Private CA Signed Certificate

If you are using a private CA, Fybrik requires a copy of the CA certificate which is used by connector/manager to validate the connection to the manager/connectors.

For each component copy the CA certificate into a file named cacerts.pem and use kubectl to create the tls-ca secret in the fybrik-system namespace.

```bash
kubectl -n fybrik-system create secret generic tls-ca \
  --from-file=cacerts.pem=./cacerts.pem
```

#### Update Values.yaml file

To use TLS the infomation about the secrets above should be inserted to the fields in [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/values.yaml) file upon Fybrik deployment using helm.

Here is an exmaple of the tls related fields in the manager that are filled based on the secrets created above:

```bash
# Manager component
manager:
  tls:
    certs:
      # Name of kubernetes secret that holds the manager certificate.
      # The secret should be of `kubernetes.io/tls` type.
      certSecretName: "tls-manager-certs"
      # Name of kubernetes tls secret namespace that holds the manager certificate.
      # The secret should be of `kubernetes.io/tls` type.
      certSecretNamespace: "fybrik-system"
      # Name of kubernetes secret that holds the certificate authority (CA) certificates
      # which are used by the manager to validate the connection to the connectors.
      cacertSecretName: "tls-ca"
      # Name of kubernetes secret namespace that holds the certificate authority (CA)
      # certificates which are used by the manager to validate the connection to the connectors.
      cacertSecretNamespace: "fybrik-system"
```

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
