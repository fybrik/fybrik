# Enabling data-plane optimization

Fybrik takes into account data governance and hard IT config policies when building a data plane. However, it does not by default take into account IT config optimization policies (i.e., optimization goals). To enable data-plane optimization, the [Optimizer component](../concepts/optimizer.md) must be enabled.

## Enabling the optimizer
Enabling the optimizer is done by setting the `solver.enabled` property to `true` in Fybrik's Helm chart. Assuming Fybrik is already deployed, the following command can be used to enable the optimizer:
```
```bash
helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set solver.enabled=true
```

## Using a custom CSP solver
The default CSP solver is the one provided by [Google OR-Tools](https://developers.google.com/optimization). A different solver from [the list of FlatZinc-supporting solvers](https://www.minizinc.org/software.html#flatzinc) can be configured by following these steps:
1. Prepare a Docker image file containing the solver executable and the solver's dependencies (e.g., dynamically-linked libraries). The executable should be called `solver` and should be placed in the directory `/data/tools/bin` of the Docker image.
2. Upload the Docker image file to any public registry.
3. Run the following command to configure the solver (this assumes that Fybrik is already deployed):
    ```bash
    helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set solver.image=<image-of-your-solver>
    ```
