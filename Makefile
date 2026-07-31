.PHONY: cluster-up infra-apply build load migrate deploy validate cluster-down

cluster-up:
	kind create cluster --name config-service

infra-apply:
	cd terraform && terraform init && terraform apply -auto-approve

build:
	docker build -t config-service:local .

load: build
	kind load docker-image config-service:local --name config-service

migrate:
	kubectl exec -n config-service $$(kubectl get pod -n config-service -l app=config-service-postgres -o jsonpath='{.items[0].metadata.name}') \
		-- psql -U postgres -d configdb -c "CREATE TABLE IF NOT EXISTS configs (id TEXT PRIMARY KEY, host TEXT NOT NULL, port INTEGER NOT NULL, app_name TEXT NOT NULL, log_level TEXT NOT NULL DEFAULT 'INFO', updated_at TIMESTAMPTZ NOT NULL DEFAULT now());"

deploy: load
	helm upgrade --install config-service ./chart

validate:
	curl -sf http://localhost:8080/ping && echo "\nping OK"
	curl -sf -X POST http://localhost:8080/configs -H "Content-Type: application/json" \
		-d '{"id":"smoke_test","host":"localhost","port":8080,"app_name":"config-service","log_level":"INFO"}' && echo "\nupsert OK"
	curl -sf http://localhost:8080/configs/smoke_test && echo "\nget OK"

cluster-down:
	kind delete cluster --name config-service
