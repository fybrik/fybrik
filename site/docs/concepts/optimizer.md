# Optimizer

The optimizer component builds an optimal data-plane plotter for a given `FybrikApplication` custom resource, taking into account:
* Available resources (modules, clusters, storage accounts)
* Application specification (e.g., geography, data access protocol)
* Specifications of the data-sets required by the application
* Governance actions required by the data-governance policy manager
* [Configuration policies](./config-policies)
* Optimization goals

The optimizer translates all the above data into a monolith [Constraint Satisfaction Problem (CSP)](https://en.wikipedia.org/wiki/Constraint_satisfaction_problem) and solves it using a third-party CSP solver. The solver returns an optimal solution in terms of the specified optimization goals. The solution is then translated into a data-plane plotter. The plotter specifies which modules should be deployed in which clusters, using which storage accounts and which configuration. Finally, the plotter is deployed to the specified clusters, resulting in a data plane that connects the required data-sets to the application.

## Using a custom CSP solver
The Constraint Satisfaction Problem is written as a [FlatZinc model](https://www.minizinc.org/doc-latest/en/fzn-spec.html). This allows using any CSP solver that supports the FlatZinc format. Currently, the default solver is the one provided by [Google OR-Tools](https://developers.google.com/optimization). Check [this list](https://www.minizinc.org/software.html#flatzinc) for other solvers supporting FlatZinc. To configure a solver different than the default solver:
1. Prepare a Docker image file containing the solver executable and the solver's dependencies (e.g., dynamically-linked libraries). The executable should be called `solver` and should be placed in the directory `/data/tools/bin` of the Docker image.
2. Upload the Docker image file to any public registry.
3. Run the following command to configure the solver:
    ```bash
    helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set solver.image=<image-of-your-solver>
    ```

## Disabling the optimizer
In the rare case of the CSP solver failing to produce any solution (which is not due to conflicting polices), Fybrik will fall back to producing a data-plane plotter without the CSP solver, but while ignoring all optimization goals. It is also possible to disable the optimizer entirely; once again, this will result in optimization goals being ignored. To disable the optimizer:
```bash
helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set solver.enabled=false
```
