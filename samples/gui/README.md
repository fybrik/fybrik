## Description

Web UI application for use by data users to create an environment for running their indicated workload with the requested data sets, for the declared purpose. Using this application, a data user can create a new environment, or manage an existing one, while monitoring the progress of the environment's construction. As a prerequisite to environment creation, the user must provide credentials to allow the environment to access the requested data asset.

The APIs and examples of their use are as follows:
  
  Create M4DApplication:
	  curl -X POST -i http://localhost:8080/v1/dma/m4dapplication --data '{"apiVersion": "app.m4d.ibm.com/v1alpha1","kind": "M4DApplication","metadata": {"name": "unittest-notebook1"},"spec": {"selector": {"matchLabels":{"app": "notebook1"}},"appInfo": {"purpose": "fraud-detection","role": "Security"}, "data": [{"catalogID": "1c080331-72da-4cea-8d06-5f075405cf17", "dataSetID": "2d1b5352-1fbf-439b-8bb0-c1967ac484b9","ifDetails": {"protocol": "s3","dataformat": "parquet"}}]}}'
	
	Get list of M4DApplications
	  curl -X GET -i http://localhost:8080/v1/dma/m4dapplication
	
	Get specific M4DApplication:
	  curl -X GET -i http://localhost:8080/v1/dma/m4dapplication/unittest-notebook1
	
	Delete M4DApplication:
	  curl -X DELETE -i http://localhost:8080/v1/dma/m4dapplication/unittest-notebook1
	
	
	Create Credentials
	  curl -X POST -i http://localhost:8080/v1/creds/usercredentials --data '{"System": "Egaria","Name": "my-notebook","Credentials": {"username": "admin"}}'
	
	Get Credentials
	  curl -X GET -i http://localhost:8080/v1/creds/usercredentials/default/my-notebook/Egaria
	  ==> returns: "{\"username\":\"admin\"}"
	
	Delete Credentials
	  curl -X DELETE -i http://localhost:8080/v1/creds/usercredentials/default/my-notebook/Egaria



	Get Environment Info
	  curl -X GET -i http://localhost:8080/v1/env/datauserenv


## Run server locally - run vault as well as REST API server
vault server -dev

export KUBECONFIG=$HOME/.kube/config

export VAULT_AUTH=JWT
export VAULT_TTL=5h
export VAULT_USER_MOUNT=v1/sys/mounts/m4d/user_creds
export VAULT_DATASET_MOUNT=v1/sys/mounts/m4d/dataset_creds
export VAULT_TOKEN= <take from local vault environment>
export VAULT_ADDRESS=http://127.0.0.1:8200/
export VAULT_USER_HOME=m4d/user_creds/
export VAULT_DATASET_HOME=m4d/dataset_creds/
export GEOGRAPHY=US

go run m4d/samples/gui/server/main.go

## Test locally
Assuming rest server is running (see run locally)
From within samples/gui/server/datauser 
go test

## Working in a cluster
GUI is deployed in the namespace the workload is running in. This should also be your current namespace.
Set the following environment variable prior to deploying the GUI. 

```
export WORKLOAD_NAMESPACE=<namespace>
export DOCKER_HOSTNAME=<registry>
```

## Creating docker images

Backend image creation is done from the main directory of the project.

```
docker build . -t $DOCKER_HOSTNAME/$WORKLOAD_NAMESPACE/datauserserver:latest -f samples/gui/server/Dockerfile.datauserserver
```
Frontend image creation

```
cd <root>/samples/gui/front-end
npm install
Ensure that .env has a correct configuration 
export NODE_OPTIONS=--max_old_space_size=4096
rm -rf build
npm run build
docker build . -t $DOCKER_HOSTNAME/$WORKLOAD_NAMESPACE/datauserclient:latest
```
## Deployment
  ```
cd <root>>/samples/gui
./deploy.sh
```

## Run 

```
kubectl port-forward service/datauserserver 8080:8080&
kubectl port-forward service/datauserclient 3000:3000&

Open a browser

Connect to localhost:3000

