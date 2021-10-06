# Dashboard sample

This use-case demonstrates how Fybrik uses runtime evaluation of policies 
to control access privileges for multiple assets.

## Mockup version of the real scenario

This is based on the smart manufacturing scenario of *Fogprotect*, 
where there exists a manufacturing factory which can be controlled/supervised via 
transmitted commands, without the need of the staff member to be physically present 
in the factory.  
The scenario here is simplified for the sake of the demo. The resources in this demo are:  
1. A robot in the manufacturing area that can be started/stopped.  
2. Two factory sectors, a production and a non-production sector, where some employees wear helmets 
and some don't. There's a video camera with image processing which monitors and counts the number of employees 
wearing/not wearing helmets in each sector. This is called the safety data of the factory.  

There are multiple roles in this scenario:  
1. A *Foreman* who is allowed to access all of the assets because he is the foreman of the factory. 
Furthermore, only the Foreman is allowed to control the robot in the manufacturing area.  
2. A *Worker* who is not allowed to control the robot, and doesn't have privileges to see the number 
of employees wearing/not wearing helmets in each of the available sectors. However, a Worker can see the total 
number of employees in each sector.  
3. An *HR* who is also not allowed to control the robot, but has access to view the number of employees 
wearing/not wearing helmets in each sector.

## Architecture
The project contains 3 main components:  
- A backend data server, responsible for reading/writing data (possibly to a database).  
- A proxy server, responsible for intercepting HTTP requests sent
from a user trying to read/write data.  
- A frontend GUI/dashboard, which helps the user perform the HTTP 
requests.

For a more detailed description of the implementation visit [fogProtect-dashboard-sample](https://github.com/fybrik/fogProtect-dashboard-sample/tree/chart-push).

## Before you begin

- Install Fybrik using the [Quick Start](../get-started/quickstart.md) guide.
- A web browser.

## Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace:

```bash
kubectl create namespace fogprotect
kubectl config set-context --current --namespace=fogprotect
```

This enables easy [cleanup](#cleanup) once you're done experimenting with the sample.

## Backend data server

The backend data server is responsible for reading/writing data to a database or any other resource. In this 
example the backend data server is a simple server that returns random numbers whenever a read request is 
received.  
Run the backend data server:  
```bash
export HELM_EXPERIMENTAL_OCI=1
helm chart pull ghcr.io/fybrik/backend-server-chart:v0.0.1
helm chart export --destination=./tmp ghcr.io/fybrik/backend-server-chart:v0.0.1
helm install rel1-backend-server ./tmp/backend_server
```

