

# This target deploys the stable crossplane chart into the kind cluster
crossplane.deploy: $(HELM)
	@$(INFO) Deploying Crossplane
	@$(HELM) repo add crossplane-stable https://charts.crossplane.io/stable
	@$(HELM) repo update
	@$(HELM) upgrade --install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace\
                     --wait --timeout 600s

# Composite target to deploy crossplane, the official grafana provider and some grafana resources
official-grafana:
	@$(MAKE) crossplane.deploy
	@$(MAKE) official-grafana.deploy
	@$(MAKE) official-grafana.resources

# Deploys the official grafana provider
official-grafana.deploy: $(KUBECTL)
	@$(INFO) Deploying official Grafana Provider
	@$(KUBECTL) apply -R -f hack/resources/provider-grafana.yaml
	@$(INFO) Waiting for the pod to be ready
	@while ! $(KUBECTL) wait --for=condition=ready -n crossplane-system pod -l pkg.crossplane.io/provider=provider-grafana --timeout=600s 2>/dev/null; do \
		echo "Pod is not ready yet, waiting..."; \
		sleep 10; \
	done
	@$(INFO) Provider Grafana is ready, deploying config
	@$(KUBECTL) apply -f hack/resources/provider-grafana-config.yaml
	@DOCKER_OS=$$(docker version -f '{{.Server.Os}}'); \
	 if [ "$$DOCKER_OS" = "windows" ] || [ "$$DOCKER_OS" = "mac" ]; then \
	   $(KUBECTL) apply -f hack/resources/provider-grafana-service-windows-mac.yaml; \
	 else \
	   $(KUBECTL) apply -f hack/resources/provider-grafana-service-linux.yaml; \
	 fi

# Deploys grafana resources
official-grafana.resources: $(KUBECTL)
	@$(INFO) Deploying Grafana resources
	@$(KUBECTL) apply -f hack/resources/organization.yaml
	@$(KUBECTL) apply -f hack/resources/datasource.yaml
	@$(KUBECTL) apply -f hack/resources/folder.yaml
	@$(KUBECTL) apply -f hack/resources/dashboard.yaml
	@$(INFO) Waiting for the resources to be ready
	@$(KUBECTL) wait --for=condition=ready organization.oss.grafana.crossplane.io example-organization --timeout=600s
	@$(KUBECTL) wait --for=condition=ready datasource.oss.grafana.crossplane.io datasource --timeout=600s
	@$(KUBECTL) wait --for=condition=ready dashboard.oss.grafana.crossplane.io dashboard --timeout=600s
	@$(KUBECTL) wait --for=condition=ready folder.oss.grafana.crossplane.io folder --timeout=600s
	@$(KUBECTL) get -n crossplane-system organization.oss.grafana.crossplane.io example-organization -o yaml > hack/resources/observations/organization.yaml
	@$(KUBECTL) get -n crossplane-system datasource.oss.grafana.crossplane.io datasource -o yaml > hack/resources/observations/datasource.yaml
	@$(KUBECTL) get -n crossplane-system folder.oss.grafana.crossplane.io folder -o yaml > hack/resources/observations/folder.yaml
	@$(KUBECTL) get -n crossplane-system dashboard.oss.grafana.crossplane.io dashboard -o yaml > hack/resources/observations/dashboard.yaml


# Cleans up the official grafana provider-config, without deleting the provider. Deleting the provider would mark all
# resources as deleted, which would then be deleted by our own provider
clean-official-grafana:
	@$(INFO) Cleaning up official Grafana
	@DOCKER_OS=$$(docker version -f '{{.Server.Os}}'); \
	 if [ "$$DOCKER_OS" = "windows" ] || [ "$$DOCKER_OS" = "mac" ]; then \
	   $(KUBECTL) delete --ignore-not-found -f hack/resources/provider-grafana-service-windows-mac.yaml; \
	 else \
	   $(KUBECTL) delete --ignore-not-found -f hack/resources/provider-grafana-service-linux.yaml; \
	 fi
	@$(KUBECTL) delete --wait=false --ignore-not-found -f hack/resources/provider-grafana-config.yaml
	@$(KUBECTL) patch -n crossplane-system providerconfig.grafana.crossplane.io default -p '{"metadata":{"finalizers":[]}}' --type=merge
	#@$(KUBECTL) delete --ignore-not-found -f hack/resources/provider-grafana.yaml

# deploys our grafana provider to test how it handles pre-existing resources from the official grafana provider
custom-grafana.deploy:
	@$(INFO) "Deploying Grafana Provider (from https://marketplace.upbound.io/account/argannor-oss/provider-grafana)"
	@$(KUBECTL) apply -R -f hack/resources/provider-grafana-this.yaml
	@$(INFO) Waiting for the pod to be ready
	@while ! $(KUBECTL) wait --for=condition=ready -n crossplane-system pod -l pkg.crossplane.io/provider=provider-grafana --timeout=600s 2>/dev/null; do \
		echo "Pod is not ready yet, waiting..."; \
		sleep 10; \
	done
	@$(INFO) Provider Grafana is ready, deploying config
	@$(KUBECTL) apply -f hack/resources/provider-grafana-this-config.yaml
	@DOCKER_OS=$$(docker version -f '{{.Server.Os}}'); \
	 if [ "$$DOCKER_OS" = "windows" ] || [ "$$DOCKER_OS" = "mac" ]; then \
	   $(KUBECTL) apply -f hack/resources/provider-grafana-service-windows-mac.yaml; \
	 else \
	   $(KUBECTL) apply -f hack/resources/provider-grafana-service-linux.yaml; \
	 fi

# runs grafana in a container on the host
grafana:
	docker run --rm --name grafana -d -p 3000:3000 grafana/grafana