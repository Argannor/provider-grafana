# Observations

These resources are observed by running the [official grafana provider](https://github.com/grafana/crossplane-provider-grafana)
and exporting the results of the resources using `kubectl get <resource> -o yaml`. We can use these to ensure that our
provider is compatible with the official provider.

To run the official grafana provider using kind use:

```bash
# Remove the current kind cluster
make dev-clean
# Create a new kind cluster
make cluster
# Run grafana on your host using docker
make grafana
# Run the official grafana provider
make official-grafana
```
