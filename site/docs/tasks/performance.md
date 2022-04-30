# Performance

When using many fybric applications at the same time the custom resource operations may take some time. This is due to the default 
concurrency of controllers being one at a time and the Kubernetes client being rate limited by default.
In order to increase the parallelism there are multiple parameters that can be controlled.

Each controller parallelism (for each fybrik custom resource) can be controlled separately. When increasing this number it's highly 
recommended to also increase the managers Kubernetes client QPS and Boost settings so that the controller won't be limited
by the amount of queries it can execute to the Kubernetes API.

An adapted helm values configuration looks like the following:
```
# Manager component
manager:
  extraEnvs:
  - name: APPLICATION_CONCURRENT_RECONCILES
    value: "5"
  - name: BLUEPRINT_CONCURRENT_RECONCILES
    value: "20"
  - name: PLOTTER_CONCURRENT_RECONCILES
    value: "2"
  - name: CLIENT_QPS
    value: "100.0"
  - name: CLIENT_BURST
    value: "200"
```

Please notice that QPS is a float while the other values are integer values.