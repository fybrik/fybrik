# Optimizer

The optimizer component builds an optimal data-plane plotter for a given `FybrikApplication` custom resource, taking into account:
* Available resources (modules, clusters, storage accounts)
* Application specification (e.g., geography, data access protocol)
* Specifications of the datasets required by the application
* Governance actions required by the data-governance policy manager
* [IT Configuration policies](./config-policies), including optimization goals

The optimizer translates all the above inputs into a monolith [Constraint Satisfaction Problem (CSP)](https://en.wikipedia.org/wiki/Constraint_satisfaction_problem) and solves it using a third-party CSP solver. The solver returns an optimal solution in terms of the specified optimization goals. The solution is then translated into a plotter. The plotter specifies which modules should be deployed in which clusters, using which storage accounts and which configuration. It also describes how data flows between the modules. Finally, the plotter is deployed to the specified clusters (via cluster-specific blueprints), resulting in a data plane that connects the required datasets to the application.

**Note:** The optimizer component is currently disabled by default, meaning all optimization goals are being ignored. Enabling it is simple and is explained [here](../tasks/data-plane-optimization.md#enabling-the-optimizer). Also note that in the rare case of the CSP solver failing to produce any solution (which is not due to conflicting polices), Fybrik will fall back to producing a plotter without the CSP solver, but while ignoring all optimization goals.

The Constraint Satisfaction Problem is written as a [FlatZinc model](https://www.minizinc.org/doc-latest/en/fzn-spec.html). This allows using any CSP solver that supports the FlatZinc format. Currently, the default solver is the one provided by [Google OR-Tools](https://developers.google.com/optimization). Check [this list](https://www.minizinc.org/software.html#flatzinc) for other solvers supporting FlatZinc. Configuring a solver different than the default solver is explained [here](../tasks/data-plane-optimization.md#using-a-custom-csp-solver).
