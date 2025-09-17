

ENV_FILE = ./config.env
DC = docker-compose --env-file $(ENV_FILE)
APP_URL = http://localhost:8080

up:
	$(DC) up --build


up-detached:
	$(DC) up -d --build


down:
	$(DC) down


down-clean:
	$(DC) down -v


build:
	$(DC) build


logs:
	$(DC) logs -f app


psql:
	docker exec -it db-itk-rest psql -U $$(grep DB_USER $(ENV_FILE) | cut -d'=' -f2) -d $$(grep DB_NAME $(ENV_FILE) | cut -d'=' -f2)


load-test:
	@echo "Запускаем нагрузочное тестирование..."
	hey -n 10000 -c 200 -m POST \
	 -H "Content-Type: application/json" \
	 -d '{"walletId":"550e8400-e29b-41d4-a716-446655440000","operationType":"DEPOSIT","amount":100}' \
	 $(APP_URL)/api/v1/wallet


export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=dude
export DB_PASSWORD=duck
export DB_NAME=itkdb

test-coverage:
	go test -coverprofile=coverage.out ./internal/wallet
	go tool cover -html=coverage.out
