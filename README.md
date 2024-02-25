# provider-grafana

`provider-grafana` is an alternative to the [official grafana provider](https://github.com/grafana/crossplane-provider-grafana).
We encountered scalability issues with the terraform based provider and decided to create a custom provider using the 
[grafana go api client](https://github.com/grafana/grafana-openapi-client-go). While we aim to be compatible with the
official one, there are some differences in the resources we support:

- ProviderConfig differs, as we don't use a json inside a secret but instead fields inside the CRD
- Currently only `Organization`, `DataSource`, `Folder`, and `Dashboard` are supported
- Only the `oss.grafana.crossplane.io` API group is supported

Use this at your own risk!

## Migrating from the official provider

Make sure to follow these steps in order, as diverging from them may result in data loss.

- Ensure you have backups of your Kubernetes resources and your Grafana database
- `kubectl delete --wait=false providerconfig.grafana.crossplane.io <name-of-your-provider-config>`
- `kubectl patch providerconfig.grafana.crossplane.io <name-of-your-provider-config> -p '{"metadata":{"finalizers":[]}}' --type=merge`
- `kubectl patch provider.pkg.crossplane.io <name-of-your-provider> -p '{"spec":{"package":"xpkg.upbound.io/argannor-oss/provider-grafana:v0.0.1-9.g4866768"}}' --type=merge`
- Wait until the provider is up and running before applying the new `ProviderConfig` (check the [example](examples/provider/config.yaml))
- Verify that the resources are still there and working as expected

## Build

Initially follow these steps:

1. Run `make submodules`
2. Run `make reviewable dev`

For subsequent builds, just run `make dev-redeploy`.


## Developing

1. Use this repository as a grafana to create a new one.
1. Run `make submodules` to initialize the "build" Make submodule we use for CI/CD.
1. Rename the provider by running the following command:
   ```shell
     export provider_name=MyProvider # Camel case, e.g. GitHub
     make provider.prepare provider=${provider_name}
   ```
1. Add your new type by running the following command:
   ```shell
     export group=sample # lower case e.g. core, cache, database, storage, etc.
     export type=MyType # Camel casee.g. Bucket, Database, CacheCluster, etc.
     make provider.addtype provider=${provider_name} group=${group} kind=${type}
   ```
1. Replace the *sample* group with your new group in apis/{provider}.go
1. Replace the *mytype* type with your new type in internal/controller/{provider}.go
1. Replace the default controller and ProviderConfig implementations with your own
1. Run `make reviewable` to run code generation, linters, and tests.
1. Run `make build` to build the provider.

Refer to Crossplane's [CONTRIBUTING.md] file for more information on how the
Crossplane community prefers to work. The [Provider Development][provider-dev]
guide may also be of use.

[CONTRIBUTING.md]: https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md
[provider-dev]: https://github.com/crossplane/crossplane/blob/master/contributing/guide-provider-development.md
