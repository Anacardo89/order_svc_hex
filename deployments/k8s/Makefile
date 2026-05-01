CALICO_VERSION := v3.27.3
CLUSTER_NAME := order-system-cluster

.PHONY: cluster-up cluster-down load-imgs copy-configs gen-secrets

cluster-up:
	@if [ "$$(kind get clusters | grep -w $(CLUSTER_NAME))" = "$(CLUSTER_NAME)" ]; then \
		echo "⚠️  Cluster '$(CLUSTER_NAME)' already exists. Skipping creation."; \
	else \
		echo "🚀 Creating kind cluster: $(CLUSTER_NAME)..."; \
		kind create cluster --name $(CLUSTER_NAME) --config kind-config.yaml; \
		echo "🔌 Installing Calico CNI..."; \
		kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/$(CALICO_VERSION)/manifests/tigera-operator.yaml; \
		kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/$(CALICO_VERSION)/manifests/custom-resources.yaml; \
		echo "🔌 Installing Ingress NginX..."; \
		kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml; \
		echo "⏳ Waiting for Ingress NGINX to be ready..."; \
		kubectl wait --namespace ingress-nginx \
		--for=condition=ready pod \
		--selector=app.kubernetes.io/component=controller \
		--timeout=90s; \
		echo "⏳ Waiting for nodes to be Ready..."; \
		kubectl wait --for=condition=Ready nodes --all --timeout=300s; \
		echo "✅ Cluster is ready!"; \
	fi

cluster-down:
	@echo "🗑️  Deleting cluster '$(CLUSTER_NAME)' and all its resources..."
	@kind delete cluster --name $(CLUSTER_NAME)
	@echo "✨ All nodes, networking, and Calico components have been removed."

load-imgs:
	@echo "🚀 Building container images..."
	@docker build -t order-api:local -f ../../order_api/Dockerfile ../../order_api
	@docker build -t order-svc:local -f ../../order_svc/Dockerfile ../../order_svc
	@echo "✅ Images built successfully!"
	@echo "🔌 Loading images into kind cluster: $(CLUSTER_NAME)..."
	@kind load docker-image order-api:local --name $(CLUSTER_NAME)
	@kind load docker-image order-svc:local --name $(CLUSTER_NAME)
	@echo "✅ Images loaded successfully!"

copy-configs:
	@echo "🏗️ Preparing deployment context..."
	@echo "📂 Copying config Files..."
	@cp ../../configs/order-api-config.yaml ./app/order-system/order_api/config.yaml
	@cp ../../configs/order-svc-config.yaml ./app/order-system/order_svc/config.yaml
	@echo "✅ Configs copied."

gen-secrets:
	@echo "📂 Initializing .env files from samples..."
	@for sample in $$(find . -name "sample.env"); do \
		target="$${sample%/*}/.env"; \
		[ -f "$$target" ] && { echo "  ⚠️ Skipping $$target"; continue; }; \
		echo "  -> Creating $$target"; \
		cp "$$sample" "$$target"; \
	done
	@echo "✅ Environment files initialized."
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
