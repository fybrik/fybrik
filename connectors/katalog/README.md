# Katalog

A data catalog and credentials manager powered by Kubernetes resources:
- [`Asset`](docs/README.md#asset) CRD for managing data assets
- `Secret` resources for managing data access credentials

## Usage

An [`Asset`](docs/README.md#asset) CRD includes a reference to a credentials `Secret`, connection information, and other metadata such as columns and associated security tags. Apply it like any other Kubernetes resource. 

Access credenditals are stored in Kubernetes `Secret` resources. You can use [Basic authentication secrets](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret) or [Opaque secrets](https://kubernetes.io/docs/concepts/configuration/secret/#opaque-secrets) with the following keys:
<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessKey</b></td>
        <td>string</td>
        <td>Access key also known as AccessKeyId</td>
        <td>false</td>
      </tr><tr>
        <td><b>secretKey</b></td>
        <td>string</td>
        <td>Secret key also known as SecretAccessKey</td>
        <td>false</td>
      </tr><tr>
        <td><b>apiKey</b></td>
        <td>string</td>
        <td>API key used in various IAM enabled services</td>
        <td>false</td>
      </tr><tr>
        <td><b>password</b></td>
        <td>string</td>
        <td>Password for basic authentication</td>
        <td>false</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>Username for basic authentication</td>
        <td>false</td>
      </tr></tbody>
</table>

## Manage users

Kubernetes RBAC is used for user management:
* To view  `Asset` resources a Kubernetes user must be granted the `katalog-viewer` cluster role. 
* To manage `Asset` resources a Kubernetes user must be granted the `katalog-editor` cluster role.

As always, create a `RoleBinding` to grant these permissions to assets in a specific namespace and a `ClusterRoleBinding` to grant these premissions cluster wide.

## Develop, Build and Deploy

The source of the `Asset` CRD are the files in the [`manifests`](manifests) directory. After modifying them run `make generate`.

Build and push the connector image with `make all` (cleanup with `make clean`).

Install with Helm as part of the standard Mesh for Data installation:
- [m4d-crd](https://github.com/IBM/the-mesh-for-data/tree/master/charts/m4d-crd) Helm chart 
  ```
  helm install m4d-crd charts/m4d-crd
  ```
- [m4d](https://github.com/IBM/the-mesh-for-data/tree/master/charts/m4d) Helm chart with `katalogConnector.enabled=true` (default).
  ```
  helm install m4d charts/m4d-crd
  ```

## Where is this going?

The current `Asset` specification was directly imported from the existing connectors API (the proto definitions) without any thought of whether this specification is the right one to use. Moving forward the entire connectors API should be refined to avoid hardcoding and all structures should be reviewed.

The plan is to experiment and check if OpenAPI 3.0 documents can be used as the core mechanism for taxonomies in Mesh for Data. The role of Katalog is to be a catalog and credentials connector that is auto generated from a reference taxonomy. The work on taxonomies is in very early stages, see https://github.com/IBM/the-mesh-for-data/issues/238.

