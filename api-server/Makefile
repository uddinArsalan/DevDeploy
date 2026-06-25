include .env
export

run:
	go run ./cmd/api/main.go

worker/run:
	go run ./cmd/worker/

rabbitmq:
	docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4-management

MIGRATE_PATH=./migrations

.PHONY: migrate_up migrate_down run worker/run

migrate_up:
	migrate -path $(MIGRATE_PATH) -database "$(DB_CONN_STR)" up

migrate_down:
	migrate -path $(MIGRATE_PATH) -database "$(DB_CONN_STR)" down

migrate_force:
	migrate -path $(MIGRATE_PATH) -database "$(DB_CONN_STR)" force $(VERSION)