# Dashboard sample

This sample application shows how Fybrik enables a dashboard application to display data from a 
backend REST server, utilizing policy-driven data protection to control what information a logged in user (role) can access.

Specifically, in this sample we demonstrate:
- Runtime policy decisions 
- A Fybrik module that supports REST protocol with JWT authentication.

Note: At this time, full integration of this application into the Fybrik framework is pending future changes to Fybrik, however this code as provided is fully operational. 

## About the dashboard application
The dashboard application here is based on the Smart Manufacturing use-case of the European 
Horizon 2020 research project [Fogprotect](https://fogprotect.eu/), based on a scenario created by BotCraft, Nagarro and Nokia in Nokia’s “Factory in a Box” technology demonstrator.  Details for the use case are available [here](https://fogprotect.eu/results/#use-cases).  

The dashboard included in this sample is a mockup of the actual dashboard which can be used to both control a manufacturing robot inside a 
"factory-in-a-box" shipping container, as well as report various metrics occuring within the box, such as the number of people detected by a video camera in the box, identifying how many of them are wearing a safety helmet. 

The scenario here is simplified for the purpose of this sample and demonstrates the following story:
1. A manufacturing robot operating inside the container which can be activated or stopped remotely by an authorized user.
2. A dashboard reporting on activity inside of the container which is divided into
a production and a non-production area.  Both areas are observed by a video camera with image 
processing that monitors and counts the number of employees with and without helmets 
in each area. This is called the safety data of the factory.  
 
In this scenario we show the following roles:  
1. A *Foreman* who is allowed to access all reported data. 
Furthermore, only the Foreman is allowed to control the robot in the manufacturing area.  
2. A *Worker* who is not allowed to control the robot, and doesn't have privileges to see the number 
of employees wearing/not wearing helmets in each of the available areas. However, a Worker can see the total 
number of employees in each area.  
3. *HR personnel* who are also not allowed to control the robot, but have access to view the number 
of employees wearing/not wearing helmets in each area.
This application demonstrates Fybrik's abilities to control what data can be seen by each role.

## Dashboard sample architecture
The project contains 3 main components:  
- A backend data service, which provides the mock container data for the dashboard.  
- A [fybrik module](https://github.com/fybrik/fogProtect-dashboard-sample/tree/main/rest-read-module), 
responsible for intercepting HTTP requests sent from a user trying to read or write data.  
- A dashboard application, which performs HTTP requests and displays the responses for the user.  While it is out of the scope of this sample application to actually provide a Java Web Token (JWT) for user authentication of the dashboard login, we emulate this step by generating role-based queries with the JWT embedded in the header. 

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
example the backend data server is a simple server that returns the mock container data.    
Run the backend data server:  
```bash
helm chart pull ghcr.io/fybrik/backend-server-chart:v0.0.1
helm chart export --destination=./tmp ghcr.io/fybrik/backend-server-chart:v0.0.1
helm install rel1-backend-server ./tmp/backend_server
```

## Register the assets

In this example we use three of the endpoints that the backend data server exposes. For each endpoint, we define 
a Fybrik *Asset* describing the data that will be returned as a response from the backend data server. In particular, the description of the Asset characterizes the type of data returned by that endpoint, as well as describes which fields within the data contain sensitive data.
This description will be used when policies are applied to the data.  

Register the assets:
```bash
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_get_safety_data.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_start_robot.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/assets/asset_stop_robot.yaml
```  
The identifier (name) of the asset reflects the URL of the end point, and is of the form: `<namespace>/<name>`, where backslash characters ("/") in the URL are replaced by a period (".") in the asset name.
For example, the name of the asset in file asset_start_robot.yaml, represented by the URL, <HOST>/api/control/start-robot, is 
`fogprotect/api.control.start-robot`.  

## Create the JWT authentication key

As the HTTP requests from the dashboard should contain the role of the logged in user in the header, the dashboard application uses JWT to encode
the relevant role in the header. The JWT is authenticated using a secret key that we store as a secret 
in the cluster, in both `fogprotect` and `fybrik-blueprints` namespaces.  
Create the JWT secret:  
```bash
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/secrets/jwt_key_secret.yaml
kubectl apply -n fybrik-blueprints -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/secrets/jwt_key_secret.yaml
```

## Fybrik Manager RBAC

In order for the Fybrik Control Plane to be able to access the JWT key that we created, give 
the Manager the relevant [RBAC authorization](https://github.com/fybrik/fogProtect-dashboard-sample/blob/main/fybrik-system-manager-rbac.yaml):  
```bash
kubectl apply -n fybrik-system -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/fybrik-system-manager-rbac.yaml
```  
## Deploy the module

In our case, the module as we described earlier is responsible for intercepting the HTTP requests received 
from the user. Once a request is received a decision must be made regarding the request based on the role of the user and the characterization of the data being requested; it should be either allowed, potentially with fields redacted, or blocked. The decision is made using 
[OpenPolicyAgent](https://www.openpolicyagent.org), and applying the policy specified 
[here](https://github.com/fybrik/fogProtect-dashboard-sample/blob/main/python/fogprotect-policy.yaml).  
    
Deploy the Fybrik module:
```bash
kubectl apply -n fybrik-system -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/rest-read-module.yaml
kubectl apply -f https://raw.githubusercontent.com/fybrik/fogProtect-dashboard-sample/main/rest-read-application.yaml
kubectl wait --for=condition=ready --all pod --timeout=120s
```
Create a port-forwarding to the new service that will be receiving the HTTP requests:  
```bash
kubectl -n fybrik-blueprints port-forward svc/rest-read 5559:5559 &
```
## Deploy the sample application  

We now deploy the dashboard that will display a table containing the safety data of the factory, along with two 
buttons to emulate starting and stopping the manufacturing robot. One can change the role of the user using a pull down menu.   
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
## Run the sample application
 
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
