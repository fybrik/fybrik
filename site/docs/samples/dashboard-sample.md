# Dashboard sample

This sample shows how Fybrik enables a dashboard application to display data from a 
backend REST server and displaying only compliant information with the person operating the dashboard.
In the sample we show how a Fybrik module that support REST protocol can control what 
information is displayed in the dashboard through policies.

Specifically in this sample we demonstrate:
- Just in time policy decisions 
- A Fybrik module that supports REST API 

Note: At this time some features in this demo use mocks as some underlying support in 
Fybrik is still pending.

## About the dashboard application
The dashboard application here is based on the smart manufacturing use-case of the European 
Horizon 2020 research project [Fogprotect](https://fogprotect.eu/). 
Details for the use case are available [here](https://fogprotect.eu/results/#use-cases).  
The dashboard included in this sample is a mock dashboard of a remote manufacturing factory 
which can be controlled/supervised by expert operators from remote, 
without the need for the expert operator to be physically present in the factory.

The scenario here is simplified for the purpose of this sample and demonstrates the following scenario:
1. A manufacturing robot that operates the manufacturing facility which can be activated or stopped remotely.
2. A safety dashboard observing the manufacturing floor which is composed of two areas. A 
production and a non-production area, that are observed by a video camera with image 
processing which monitors and counts the number of employees with and without helmets 
in each area. This is called the safety data of the factory.  

We demonstrate Fybrik abilities to control what data is seen by each role. 
In this scenario we show the following roles:  
1. A *Foreman* who is allowed to access all of the assets. 
Furthermore, only the Foreman is allowed to control the robot in the manufacturing area.  
2. A *Worker* who is not allowed to control the robot, and doesn't have privileges to see the number 
of employees wearing/not wearing helmets in each of the available areas. However, a Worker can see the total 
number of employees in each area.  
3. *HR personnel* who are also not allowed to control the robot, but have access to view the number 
of employees wearing/not wearing helmets in each area.

## Dashboard sample architecture
The project contains 3 main components:  
- A backend data service, which provides the mock data for the dashboard.  
- A [fybrik module](https://github.com/fybrik/fogProtect-dashboard-sample/tree/main/rest-read-module), 
responsible for intercepting HTTP requests sent from a user trying to read or write data.  
- A dashboard application, which performs HTTP requests and displays the responses for the user.

For a more detailed description of the implementation visit [fogProtect-dashboard-sample](https://github.com/fybrik/fogProtect-dashboard-sample/tree/main).

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

The backend data server is responsible for reading or writing data to a database or any other resource. In this 
example the backend data server is a simple server that returns the mock data for the dashboard.    
Run the backend data server:  
```bash
helm chart pull ghcr.io/fybrik/backend-server-chart:v0.0.1
helm chart export --destination=./tmp ghcr.io/fybrik/backend-server-chart:v0.0.1
helm install rel1-backend-server ./tmp/backend_server
```

## Register the assets

In this example we use 3 of the endpoints that the backend data server exposes. For each endpoint, we define 
an asset describing the data that will be returned as response from the backend data server. The description is used 
later to apply policies on the data.  
Register the assets:
```bash
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_get_safety_data.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_start_robot.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_stop_robot.yaml
```  
The identifier of the assets is `<namespace>/<name>`, and it should be used in the `FybrikApplication` that we will 
describe later on. For example the identifier of the asset defined in the file `asset_start_robot.yaml` is 
`fogprotect/api.control.start-robot`.  

## Create JWT authentication key

As the HTTP requests should contain the role of the user in the header, the dashboard application uses JWT to pass 
the JWT of the relevant role in the header. The JWT is authenticated using a secret key that we store as a secret 
in the cluster, in both `fogprotect` and `fybrik-blueprints` namespaces.  
Create the JWT secret:  
```bash
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/secrets/jwt_key_secret.yaml
kubectl apply -n fybrik-blueprints -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/secrets/jwt_key_secret.yaml
```

## Fybrik manager RBAC

In order for the fybrik manager to be able to access the assets and to deploy the fybrik module that we created, give 
the manager the relevant [RBAC authorization](https://github.com/fybrik/fogProtect-dashboard-sample/blob/main/fybrik-system-manager-rbac.yaml):  
```bash
kubectl apply -n fybrik-system -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/fybrik-system-manager-rbac.yaml
```  

## Deploy the module

In our case, the module as we described earlier is responsible for intercepting the HTTP requests received 
from the user. Once a request is received a decision must be made regarding the request, it should be either allowed, 
columns redacted or blocked depending on the role of the user. The decision is made using 
[OpenPolicyAgent](https://www.openpolicyagent.org), and applying the policy described in 
[About the dashboard application](#About the dashboard application) and specified 
[here](https://github.com/fybrik/fogProtect-dashboard-sample/blob/main/python/fogprotect-policy.yaml).  
Deploy the fybrik module and application:
```bash
kubectl apply -n fybrik-system -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/rest-read-module.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/rest-read-application.yaml
kubectl wait --for=condition=ready --all pod --timeout=120s
```

## Deploy the dashboard  

We now deploy the dashboard that will display a table that contains the safety data of the factory, along with two 
buttons to start and stop the manufacturing robot. One can change the role of the user using a pull down menu.  
First create a port-forwarding to the service that will be intercepting the HTTP requests:  
```bash
kubectl -n fybrik-blueprints port-forward svc/rest-read 5559:5559 &
```

Afterwards, deploy the dashboard application:  
```bash
helm chart pull ghcr.io/fybrik/factory-gui-chart:v0.0.1
helm chart export --destination=./tmp ghcr.io/fybrik/factory-gui-chart:v0.0.1
helm install rel1-factory-gui ./tmp/factory_gui
kubectl wait --for=condition=ready --all pod --timeout=120s
```

Lastly, create a port-forwarding to the dashboard service in order to be able to open the dashboard in your browser:  
```bash
kubectl port-forward svc/factory-gui 3001:3000 &
```

Open your browser and go to http://127.0.0.1:3001.

## Cleanup

1. Stop the port-forwarding:
    ```shell
    pgrep -f "kubectl port-forward svc/factory-gui 3001:3000" | xargs kill
    pgrep -f "kubectl -n fybrik-blueprints port-forward svc/rest-read 5559:5559" | xargs kill
    ```
2. Remove the `tmp` directory that was created temporarily:
    ```shell
    rm -r tmp
    ```
3. Delete the fybrik application and module:
    ```shell
    kubectl delete fybrikapplication rest-read
    kubectl -n fybrik-system delete fybrikmodule rest-read-module
    ```
4. Delete the `fogprotect` namespace:  
    ```shell
    kubectl delete namespace fogprotect
    ```
5. Delete the RBAC authorization of the manager:  
    ```shell
    kubectl -n fybrik-system delete -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/fybrik-system-manager-rbac.yaml
    ```
6. Delete the JWT secret:  
    ```shell
    kubectl delete -n fybrik-blueprints -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/secrets/jwt_key_secret.yaml
    ```
