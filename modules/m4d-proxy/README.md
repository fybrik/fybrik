# m4d-proxy

## Pre-reqs

Install standard tools:

```
make install-tools
```

Install helm3 on the local machine

```
curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```

## Environment

```
export NAME=bucket1
export BUCKET=m4d-objectstorage-secret-provider-test
export ENDPOINT=s3.eu-de.cloud-object-storage.appdomain.cloud
export APIKEY=<apikey>
```

## Cluster

```
make kind
```

## Installation

Install using the helm
```
cd modules/m4d-proxy
make install
make list
make status
cd -
```

## Testing

Testing against local service: bucket1.default.svc.cluster.local

```
cd samples/m4d-proxy
make app-install
make app-status                 # until pod is running
make test
cd -
```

