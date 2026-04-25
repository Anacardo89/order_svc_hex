CALICO_VERSION := v3.27.3
CLUSTER_NAME := order-system-cluster

.PHONY: cluster-up cluster-down gen-secrets init-envs

cluster-up:
	@if [ "$$(kind get clusters | grep -w $(CLUSTER_NAME))" = "$(CLUSTER_NAME)" ]; then \
		echo "⚠️  Cluster '$(CLUSTER_NAME)' already exists. Skipping creation."; \
	else \
		echo "🚀 Creating kind cluster: $(CLUSTER_NAME)..."; \
		kind create cluster --name $(CLUSTER_NAME) --config kind-config.yaml; \
		echo "🔌 Installing Calico CNI..."; \
		kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/$(CALICO_VERSION)/manifests/tigera-operator.yaml; \
		kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/$(CALICO_VERSION)/manifests/custom-resources.yaml; \
		echo "⏳ Waiting for nodes to be Ready..."; \
		kubectl wait --for=condition=Ready nodes --all --timeout=300s; \
		echo "✅ Cluster is ready!"; \
	fi

cluster-down:
	@echo "🗑️  Deleting cluster '$(CLUSTER_NAME)' and all its resources..."
	@kind delete cluster --name $(CLUSTER_NAME)
	@echo "✨ All nodes, networking, and Calico components have been removed."

gen-secrets:
	@echo "🔐 Deep-scanning for secret templates..."
	@for template in $$(find . -name "secret.yaml.tmpl"); do \
		dir=$$(dirname $$template); \
		if [ -f "$$dir/.env" ]; then \
			echo "  -> Processing: $$dir"; \
			( \
				cd $$dir && \
				set -a; . ./.env; set +a; \
				envsubst < secret.yaml.tmpl | kubectl apply -f -; \
			) \
		else \
			echo "  ⚠️  Skipping $$dir: Template found, but no .env file present."; \
		fi \
	done
	@echo "✅ Secret generation complete."

init-envs:
	@echo "📂 Initializing .env files from samples..."
	@for sample in $$(find . -name "sample.env"); do \
		dir=$$(dirname $$sample); \
		target="$$dir/.env"; \
		if [ ! -f "$$target" ]; then \
			echo "  -> Creating $$target from sample..."; \
			cp "$$sample" "$$target"; \
		else \
			echo "  ⚠️  Skipping $$dir: .env already exists."; \
		fi \
	done
	@echo "✅ Environment files initialized."