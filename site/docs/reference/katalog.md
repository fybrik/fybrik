# Katalog

Katalog is a data catalog that is included in Fybrik for evaluation purposes.

It is powered by Kubernetes resources:

- [`Asset`](./crds.md#asset) CRD for managing data assets
- `Secret` resources for managing data access credentials

## Usage

An [`Asset`](./crds.md#asset) CRD includes a reference to a credentials `Secret`, connection information, and other metadata such as columns and associated security tags. Apply it like any other Kubernetes resource. 

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
        <td><b>access_key</b></td>
        <td>string</td>
        <td>Access key also known as AccessKeyId</td>
        <td>false</td>
      </tr><tr>
        <td><b>secret_key</b></td>
        <td>string</td>
        <td>Secret key also known as SecretAccessKey</td>
        <td>false</td>
      </tr><tr>
        <td><b>api_key</b></td>
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

