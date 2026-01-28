.PHONY: run_server db grpc_server migration_create migrate_up migrate_down

run_server:
	go run ./cmd/entrypoint/api/main.go
	
db:
	./cmd/gen/db.sh

grpc_server:
	./cmd/gen/grpc_server.sh

# DATABASE MIGRATION
# Example: make migration_create name=init
migration_create: 
	migrate create -ext sql -dir migrations -seq $(name)

# Example: make migration_up
migration_up:
	go run ./cmd/entrypoint/migration/main.go up

# Example: make migration_down
migration_down:
	go run ./cmd/entrypoint/migration/main.go down
