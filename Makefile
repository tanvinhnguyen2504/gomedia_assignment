-include .env
export

DB_HOST   ?= localhost
DB_PORT   ?= 5432
DB_USER   ?= postgres
DB_NAME   ?= viewings_db

PSQL = psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER)

db-create:
	$(PSQL) -c "CREATE DATABASE $(DB_NAME);" postgres

db-schema:
	$(PSQL) -d $(DB_NAME) -f schema.sql

db-seed:
	$(PSQL) -d $(DB_NAME) -f seed.sql

db-init: db-create db-schema db-seed

run:
	go run main.go

dev:
	air

swagger:
	swag init --generalInfo main.go --output docs

test:
	go test ./internal/... -v -run TestService
