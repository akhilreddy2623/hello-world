INTEGRATION_TIMEOUT = 120s
project_root=../
## up: starts all containers in the background without forcing build
up:
	@echo "Starting Infrastructure Docker images..."
	docker-compose -f docker-compose-infra.yml up -d
	@echo "Docker images started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker-compose -f docker-compose-infra.yml down
	@echo "Done!"

start:
	@echo "Starting Application Docker images..."
	docker-compose -f docker-compose-app.yml up -d
	@echo "Docker images started!"

stop:
	@echo "Stopping docker compose..."
	docker-compose -f docker-compose-app.yml down
	@echo "Done!"

build_vault_api:
	@echo "Building vault api binary..."
	bash ./scripts/build/payment-vault-api-build.sh
	@echo "Done!"
build_administrator_api:
	@echo "Building administrator api binary..."
	bash ./scripts/build/payment-administrator-api-build.sh
	@echo "Done!"
build_executor_api:
	@echo "Building executor api binary..."
	bash ./scripts/build/payment-executor-api-build.sh
	@echo "Done!"
build_task_manager_api:
	@echo "Building task manager api binary..."
	bash ./scripts/build/task-manager-api-build.sh
	@echo "Done!"
build_administrator_worker:
	@echo "Building administrator worker binary..."
	bash ./scripts/build/payment-administrator-worker-build.sh
	@echo "Done!"
build_vault_worker:
	@echo "Building vault worker binary..."
	bash ./scripts/build/payment-vault-worker-build.sh
	@echo "Done!"
build_executor_worker:
	@echo "Building executor worker binary..."
	bash ./scripts/build/payment-executor-worker-build.sh
	@echo "Done!"
build_config_manager_api:
	@echo "Building config manager api binary..."
	bash ./scripts/build/config-manager-api-build.sh
	@echo "Done!"
build_data_hub_worker:
	@echo "Building data hub worker binary..."
	bash ./scripts/build/data-hub-worker-build.sh
	@echo "Done!"

build: build_executor_api build_task_manager_api build_administrator_api build_vault_api build_administrator_worker build_executor_worker build_vault_worker build_config_manager_api build_data_hub_worker

build-start: build
	@echo "Stopping docker compose..."
	docker-compose -f docker-compose-app.yml down
	@echo "Starting Application Docker images..."
	docker-compose -f docker-compose-app.yml up --build -d
	@echo "Done!"

lint: 
	cd $(project_root) && bash ./project/scripts/lint.sh
## unit test only
test:
	cd $(project_root) && bash ./project/scripts/test/test.sh

integration_all:
	cd $(project_root) && bash ./project/scripts/test/integration-test.sh | tee ./testing/integration_all.log

coverage:
	cd $(project_root) && bash ./project/scripts/test/run-coverage.sh

integration_up: build_vault_api
	@echo "Starting Infrastructure Docker images..."
	docker compose -f ./docker-compose-infra-test.yml -p payment-test up -d
	@echo "Waiting 10 seconds for the containers to be ready"
	sleep 10
	@echo "Setting up Tables and Topics!"
	cd $(project_root) && go run ./project/setup_infra.go

integration_down:
	@echo "Stopping docker compose..."
	docker compose -f ./docker-compose-infra-test.yml -p payment-test down
	@echo "Done!"

integration_administrator:
	@echo "Running administrator integration tests..."
	go test --tags=integration -v -count=1 -timeout=${INTEGRATION_TIMEOUT} ../payment-administrator/integration-tests/... | tee ../testing/integration_administrator.log
	@echo "Done!"

integration_taskmanager:
	@echo "Running taskmanager integration tests..."
	go test --tags=integration -v -count=1 -timeout ${INTEGRATION_TIMEOUT} ../task-manager/integration-tests/... | tee ../testing/integration_taskmanager.log
	@echo "Done!"

integration_executor:
	@echo "Running executor integration tests..."
	go test --tags=integration,azureEndpoint -v -count=1 -timeout ${INTEGRATION_TIMEOUT} ../payment-executor/integration-tests/... | tee ../testing/integration_executor.log
	@echo "Done!"
	