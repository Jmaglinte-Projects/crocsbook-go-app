.PHONY: run_server db

run_server:
	go run ./cmd/entrypoint/api/main.go
	
db:
	./cmd/gen/db.sh
