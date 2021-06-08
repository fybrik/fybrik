# Architecture

Mesh for Data takes a modular approach to provide an open platform for controlling
and securing the use of data across an organization. The figure below showcases the
current architecture of the Mesh for Data platform, running on top of Kubernetes. 
The storage systems shown in the lower half of the figure are merely an example.

The core parts of Mesh for Data are based on [Kubernetes operators](https://www.openshift.com/learn/topics/operators) and custom resource definitions in order to define its work items and reconcile the state of them.

The primary interaction object for a data user is the `M4DApplication` CRD where a user defines which data should be used for which purpose. The following chart and description describe the architecture and components of Mesh for Data relative to when they are used.

![Architecture](../static/workflow_multicluster.svg)

Before the data user can perform any actions a data operator has to [install](../get-started/quickstart.md) the Mesh for Data and modules. 
[Modules](./modules.md) are an extensible mechanism for processing data.
Modules can provide read/write access or produce implicit copies that serve as lower latency caches of remote assets. Modules also enforce policies defined for data assets.

Mesh for Data connects to external services to receive policy engine decisions, available data and credentials. Policies, assets and access credentials to the assets have to be defined before the user can run an application. The current abstraction supports 2 different [connectors](./connectors.md): one for data catalog and one for policy manager. It's designed in an open way so that multiple different catalog and policy frameworks of all kinds of cloud and on-prem systems can be supported. The data steward configures policies in an external policy manager over assets defined in an external data catalog. Dataset credentials are retrieved from Vault by using [Vault API](https://www.vaultproject.io/api). Vault uses a custom secret engine implemented with [HashiCorp Vault plugins system](./vault_plugins.md) to retrieve the credentials from where they are stored (data catalog for example).

Once a developer submits an `M4DApplication` CRD to Kubernetes the ApplicationController will make sure that all the specs are fulfilled and will make sure that the data is made accessible to the user according to the previously defined policies. The `M4DApplication` holds metadata about the application such the data assets required by the application, the processing purpose and the method of access the user wishes (protocol e.g. S3 or Arrow flight). 
It uses this information to check with external systems (4) if access or copy is allowed
and whether restrictive policies such as masking or hashing have to be applied. It compiles blueprints based on on the policy decisions received via the connectors and chooses the modules (5) which are best fit for the requirements that the user specified regarding the access protocol and availability.
As data assets may reside in different systems the blueprints are compiled in a `Plotter` CRD (6) that specifies which blueprints have to be executed in which cluster.

Depending on the setup the `PlotterController` will use various methods to distribute the blueprints. In a multi cluster setup the default distribution implementation is using [Razee](http://razee.io) to control remote blueprints, but several multi-cloud tools
could be used as a replacement. The `PlotterController` also collects statuses and distributes
updates of said blueprints. Once all the blueprints on all clusters are ready the plotter is marked as ready.

A single [blueprint](../reference/crds.md#blueprint) contains the specification of all assets that shall be accessed in a single cluster by a single application.
The `BlueprintController` makes sure that a blueprint can deploy all needed modules (8) and (9) and tracks their status (10). Once e.g. an implicit-copy module finishes the copy the blueprint is also in a ready state.
A read or write module is in ready state as soon as the proxy service such as the arrow-flight module is running. 

In this example an [implicit-copy module](../reference/ddc.md) copies data from a remote postgres database into a S3 compatible ceph instance.
The arrow-flight module then locally serves the data to the user via the Arrow flight protocol. Credentials are handled by the modules (11) and are never exposed to the user. The application reads from and writes data to allowed targets. 
Requests are handled by M4DModule instances(12). The application can not interact with unauthorized targets.
