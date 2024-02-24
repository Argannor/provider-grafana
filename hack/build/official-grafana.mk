


crossplane.deploy: $(HELM)
	@$(INFO) Deploying Crossplane
	@$(HELM) repo add crossplane-stable https://charts.crossplane.io/stable
	@$(HELM) repo update
	@$(HELM) upgrade --install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace\
                     --wait --timeout 600s


official-grafana.deploy: $(KUBECTL)
	@$(INFO) Deploying official Grafana Provider
	@$(KUBECTL) apply -R -f hack/resources/provider-grafana.yaml
	@$(INFO) Waiting for the pod to be ready
	@while ! $(KUBECTL) wait --for=condition=ready pod -l pkg.crossplane.io/provider=provider-grafana --timeout=600s 2>/dev/null; do \
		echo "Pod is not ready yet, waiting..."; \
		sleep 10; \
	done
	@$(INFO) Provider Grafana is ready, deploying config
	@$(KUBECTL) apply -f hack/resources/provider-grafana-config.yaml