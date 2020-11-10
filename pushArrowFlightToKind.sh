export HELM_EXPERIMENTAL_OCI=1
./hack/tools/bin/helm chart pull ghcr.io/the-mesh-for-data/arrow-flight-module-chart:latest
./hack/tools/bin/helm chart export --destination modules ghcr.io/the-mesh-for-data/arrow-flight-module-chart:latest
./hack/tools/bin/helm chart save modules/arrow-flight-module kind-registry:5000/m4d-system/arrow-flight-module-chart:latest
./hack/tools/bin/helm chart push kind-registry:5000/m4d-system/arrow-flight-module-chart:latest
