# Katalog

A data catalog and credentials manager powered by Kubernetes resources:
- `Asset` CRD for managing data assets
- `Secret` resources for managing data access credentials

## Usage

See [documentation](https://fybrik.io/latest/reference/katalog/) in the website.

## Develop, Build and Deploy

After making changes to the CRD you must run `make generate manifests` from the project's root directory.

Build and push the connector image with `make all`.

Install with Helm as part of the standard Fybrik installation:
- [fybrik-crd](https://github.com/fybrik/fybrik/tree/master/charts/fybrik-crd) Helm chart 
  ```
  helm install fybrik-crd charts/fybrik-crd
  ```
- [fybrik](https://github.com/fybrik/fybrik/tree/master/charts/fybrik) Helm chart with `katalogConnector.enabled=true` (default).
  ```
  helm install fybrik charts/fybrik-crd
  ```
