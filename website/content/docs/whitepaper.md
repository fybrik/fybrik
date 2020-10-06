# Delivering Enterprise Compliance for Data Usage - The Mesh for Data

The Mesh for Data is a cloud native platform that connect, secure, control and observe the use of data --- it orchstrates the services and facilities that involves the use of data, while addressing security and governance and ensuring that the usage of data would be complaint with the data usage policies. The Mesh for Data enables the user workload to focus on the business goals delegating concerns for governance, securiy, access and location of data to the mesh.

## The Challange

Data is critical to derive value and insight for organizations, and is a key element for businneses and goverments to efficently operate. However, easily taking advantage of data is challenging due to the need to address security, governance, discovery and lineage. More so, when deriving insight in the era of hybrid cloud.
In particular, today after access has been granted to a data set, the organization loses control of what is done with the data and it is only via manual processes that it is possible to ensure the data is used only for allowed purposes in allowed locations.

Today, processing data and generate insight in a compliant and governed manner there are three main personas involved:
* A data user such as a data scientist or analyst who is deriving business value from the data
* An operator who is responsible for managing the data repositories and the compute infrastructure
* A data steward who is responsible for defining and enforcing the policies to protect data assets and ensure compliance with regulations

Each of these personas suffers from a range of difficulties. For instance, the data user often needs to endure a long and manual process to get permission to use the data needed for the task, the operator needs to manage lots of manual copies as data is moved between clouds in a hybrid setting, and the governance officer has to ensure compliance with policies in a dynamic and global regulatory environment. As a whole, the organization suffers from the lost opportunity cost due to the time it takes to be able to use data, the risk of reputational impact if data is lost and the risk of fines and reputation for non-compliance with regulations.

## The Vision

The Mesh for Data enables the above persona to achieve the business goals by focusing on declarting and specificying what should be done and have the mesh for data enforce and reflect the required policies and regulation to ensure compliance.
Each of the persoans interact with native tools to specify how the workload should behave.
* The data user develops and runs the models, analytics or code using native tools such as notebooks and deploy the workload in containers specifying its business goals
* The operator describe the infrastructure (e.g., cloud resources, data locations, networking) using kuberentes mechanisms
* Thge data steward specify policies and observe compliance through data catalogs and policy managers

{{< image width="80%" ratio="0%" src="mesh-for-data-vision.png" >}}

The Mesh for Data is a delivered as a combination of control plane orchestration and run-time modules. The Mesh for Data interacts with data catalogs and policy management tools, and deploy run-time modules that handle data access, security and governance in an holistic manner. The Mesh for Data control plane wraps the application workload (i.e., user core business logic) with run-time modules that are responsible for intermediating between the user code and the external resource (e.g., data).

## The Concept

The Mesh for Data wraps the user workload by leveraging building blocks from Kubernetes and Istio. It enables running containarized application as-is, but intermidiates all ingress and egress flows to and from the application. Internal communication between the application containers are not affected by the mesh for data.
The mesh for data leverages a declerative specification that describe the purpose of the workload and the required external resources (i.e., data, communication) that are needed to run.
Using the above information, the Mesh for Data isolates the application workload and control communication between the application and the external world only through modules that are deployed by the mesh for data.

The Mesh for Data address the following main pillars:
* Connect: provide applications with access to data oblivious to location of either the computation or the data, including the ability to schedule compute near the data or data mobility as needed by the computation. Moreover, supporting transparently various application APIs and data source APIs.
* Secure: control access to data and deliver data to the applications taking into account the purpose of the computation, the location and the relevant policies that applies to the use of data. Moreover, control how the output of the computation is handled, communicated and stored.
* Control: orchstrate computation and data mobility based on requirements and policies.
* Observe: collection information on the use of data and provide metrics, logs, audit and data linegae for performance, governance and compliance.

## Archietcture

The Mesh for Data is composed of control plane augmented by connectors that inteacts with external services (e.g., data catalog, policy manager, credentials) that act as an orcestration layer that deploy run-time components, called modules, which enforces compliance with the policies and requirements.

### The Control Plane

The Mesh for Data control plane is a set of Kubernetes operators which are named the manager.
The manager is driven by input from the M4DApplication custrom resource defeinition (CRD)

Mesh for Data Control Plane: The Mesh for Data control plane is a component composed of a collection of Kubernetes controllers, services, connectors and registries that is used to define the way an application running on the Mesh for Data should be deployed, based upon inputs about the application to execute (provided in the M4Dapplication) and information retrieved from external policy manager, and data catalog, and data credential manager.

Mesh for Data Runtime: The run-time is composed of modules instances (M4DModules) that intermediate between the application workload and the external world (e.g., data and services). The run-time is deployed according to the blueprint, which is generated by the M4DApplication Controller.


### The Control Plane

{{< image width="80%" ratio="0%" src="mesh-for-data-control-plane.png" >}}


{{< image width="80%" ratio="0%" src="mesh-for-data-runtime.png" >}}

The 
The application runs in its native Kuberetes namespace 
The Mesh for data 
The description of the application workload is a key entity for the mesh for data, and is defined by an application description (i.e., M4DApplication resource). Before running a workload, an application needs to be registred to the Mesh for Data control plane by applying a M4DApplication resource. This is the declarative definition provided by the data user. The registration provides context about the application such as the purpose for which it’s running, the data assets that it needs, and a selector to identify the workload. Additional context such as geo-location is extracted from the platform. The actions taken by the Mesh for Data are based on data policies and the context of the application. Specifically, the Mesh for Data does not consider end-users of an application. It is the responsibility of the application to implement mechanisms such as end user authentication if required, e.g. using Istio authorization with JWT.

The new DatMesh idea takes declarative descriptions of the needs of various personas, such as a data user or a governance officer, and using a library of plugins "compiles" into a deployment description of containers and sidecars, encoded as a CRD.  This includes the user’s computation as well as injected code, which ensures governance is enforced, lineage is tracked, new data sets are registered, etc.  This talk will motivate and describe the DatMesh concept we are looking to grow with a community.


As can be seen in the figure below, which shows a  basic data mesh, the application instead of having direct access to the data, gets access to the data it needs via a plugin which knows how to enforce governance policies before the data is shipped to the application.  We use the mechanisms of Kubernetes and Istio to control and limit the applications communication.  For instance, in this example, the application container is wired such that it can only get its inputs from the container that enforces governance.  In addition, by building upon mechanisms of Kubernetes and Istio, we can measure the code that is being run and verify where it is running in order to inject the credentials to access the source data but only if the code is authentic and running in the right location.  In this way, we can ensure, for example, that data that needs to be processed in the EU is only processed on Kubenetes clusters in the EU.
In the same way we can control the input of the application, we can control what it does with its output.  In this example, we ensure that the application can only write its results to explicitly white-listed addresses, e.g., preventing writing the data to Box if not permitted.  We can also inspect and if needed modify the results in the same way we can modify the inputs.  For instance, if the application saw credit card numbers for purposes of building a fraud model, we can ensure that these credit card numbers are never leave the application.


Since the mesh is build on top of Kubernetes, it can be deployed anywhere a client has available OpenShift clusters.  And in particular, our vision is to make it  possible to deploy a single mesh in a hybrid environment, such that some of the mesh's containers run on one cluster and other containers run on a different cluster.  This will support both hybrid and multi cloud deployments.


Our approach is based on open standard and tools, such as Kubernetes, Istio, and we envision contribution parts of the core technology as open source.


Through this data mesh the business can accelerate the use of data while keeping in with regulatory compliance, as regulatory compliance would be delivered through automation instead of human interactions.


IBM is actively exploring use-cases and requirements for the data mesh.  In particular, to ensure that our technology addresses real business needs, we would like to engage with potential users, in particular in financial services, to build a demonstrator of a real application.






Background on the Mesh for Data and Key Technical Concepts and Components

The Mesh for Data is a combination of control plane orchestration tool and run-time modules that provides the ability to unify data access, security and governance in a cloud independent but native way. The Mesh for Data control plane wraps the application workload (i.e., user core business logic) with run-time modules that are responsible for intermediating between the user code and the external resource (e.g., data).

The description of the application workload is a key entity for the mesh for data, and is defined by an application description (i.e., M4DApplication resource). Before running a workload, an application needs to be registred to the Mesh for Data control plane by applying a M4DApplication resource. This is the declarative definition provided by the data user. The registration provides context about the application such as the purpose for which it’s running, the data assets that it needs, and a selector to identify the workload. Additional context such as geo-location is extracted from the platform. The actions taken by the Mesh for Data are based on data policies and the context of the application. Specifically, the Mesh for Data does not consider end-users of an application. It is the responsibility of the application to implement mechanisms such as end user authentication if required, e.g. using Istio authorization with JWT.



 ## Archietcture - Key Components


Mesh for Data Application (M4Dapplication):  This is a YAML description of the application workload that will execute on Kubernetes and be wrapped by the Mesh for Data. It provides the context about the application such as the purpose for which it’s running, the data assets that it needs, and a selector to identify the workload.  Example workloads are Watson Studio notebook, a Data Stage Job, etc., or post v1 it could be a bespoke application.

Mesh for Data Control Plane: The Mesh for Data control plane is a component composed of a collection of Kubernetes controllers, services, connectors and registries that is used to define the way an application running on the Mesh for Data should be deployed, based upon inputs about the application to execute (provided in the M4Dapplication) and information retrieved from external policy manager, and data catalog, and data credential manager.

Mesh for Data Runtime: The run-time is composed of modules instances (M4DModules) that intermediate between the application workload and the external world (e.g., data and services). The run-time is deployed according to the blueprint, which is generated by the M4DApplication Controller.

Mesh for Data Connector: Connectors are used by the control plane to interact with the workload context. Namely, the data catalog that describes the data assets, the policy manager that holds relevant policies, and credential manager that keeps data assets credentials. Connectors are gRPC services that are deployed by the Mesh for Data and are configurable.

Policy Manager Connector: A service that interacts with an external policy manager and is responsible to communicate the enforcement actions that the Mesh for Data needs to apply to the application workloads. Enforcing data governance policies requires a Policy Manager or a Policy Decision Point (PDP) that dictates what enforcement actions need to take place. The Mesh for Data supports a wide and extensible set of enforcement actions to perform on data read, write or copy. These include transformation of data, verification of the data, and various restrictions on the external activity of an application that can acceess the data.  A PDP returns a list of enforcement actions given a set of policies and specific context about the application and the data it uses. The Mesh for Data open source distribution includes an example PDP that is powered by Open Policy Agent (OPA). A Policy Compiler component handles combining enforcement actions from multiple policy manager connectors, if configured.

Data Catalog Connector: The Mesh for Data assumes the use of an enterprise data catalog. For example, to reference a required data asset in a M4DApplication resource, you provide a link to the asset in the catalog. The catalog provides metadata about the asset such as security tags. It also provides connection information to describe how to connect to the data source to consume the data. The Mesh for Data uses the metadata provided by the catalog both to enable seamless connectivity to the data and as input to making policy decisions. The data user is not concerned with any of it and just selects the data that s/he needs regardless of where the data resides.  The Mesh for Data does not contain a data catalog. Instead, it links to external data catalogs using connectors. 

Credentials Manager Connector: The Mesh for Data assumes that data access credentials of cataloged assets are securely stored in an enterprise credentials manager and are referenced from the data catalog. The Mesh for Data links to existing credential managers using connectors for reading data access credentials. The Mesh for Data contains an example connector to HashiCorp Vault.


M4DApplication Controller: Responsible for compiling a blueprint that describe compliant data flows for the application workload. It processes the Mesh for Data application YAML and retrieves a set of modules to implement all governance and management actions based upon defined policies retrieved by the Policy Compiler service. Governance actions are re-generated upon any policy change. The controller etrives information about the resources and modules that it can use. Specifically, it searches for modules that can satisfy the workload requirements (e.g., governance enforcement actions) taking into account performance considerations, the use of specific client SDKs and more. The controller generates a Blueprint describing the entire data path, including components injected into the data ingress and egress paths of the application.

Compute and Storage Resource Registry: The registries that are used by the Mesh for Data to understand the available infrastructure to the workload. These registries are not yet implemented.

Mesh for Data Module:  Modules are the run-time components of the Mesh for Data and allow controlling and intermediating the flow of information for the application workload. Modules can be open or proprietary and are defined in a registry. Modules are the way to package function such as injectable services, jobs, and configurations and make them available to the control plane to choose from, and to be deployed during run-time.



Blueprint: The M4DApplication Controller builds a blueprint for each M4Dapplication applied and received by the control plane. The blueprint describes the information flows and modules for the workload, and includes services, jobs, and configurations that are injected into the data path by the control plane to fulfill security, governance, data management, and application requirements.  

Blueprint Controller: A k8s orchestrator that processes the blueprint and deploys all runtime components accordingly.  


Standard capabilities implemented for all M4Dapplication instances:

JWT Token Injector: The Mesh for Data includes a module that knows how to add a JWT token into a REST request that the application sends. The token injector retrieves the proper token from the credential connector and automatically adds it to the proper REST requests.

Isolation and Gateway Controller: Responsible for controlling the external interactions of the application workload. All communication from application workload pods to non-workload pods is controlled by this module. All traffic to external endpoints is routed through a gateway that can control and intermediate those flows. 


Modules used depending on M4Dapplication information, data types, and governance requirements:

Implicit Copy Controller:   The Mesh for Data uses this module to perform an implicit copy of a data set from one data store to another with possibly some transform applied to the data set, e.g., for governance.  This is a transient persistent copy that is clean up when the application completes execution.  This module will be inserted into the blueprint by the M4DApplication Controller depending upon the location of the data set, the location where the application will run and governance requirements.

Apache Arrow Flight Data Access Module: A Mesh for Data module that exposes an Arrow/Flight interface and is able to apply transformations on data accessed through the module.  This module is used to support governance on data that is directly accessed without an implicit copy.
