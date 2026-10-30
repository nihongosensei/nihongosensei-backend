.PHONY: dev test

# Load local configuration (ELASTICSEARCH_*, LOG_LEVEL, PORT, ...) if present.
-include .env
export

# Run the dev server (listens on $PORT, default 3001).
dev:
	go run ./cmd/server

test:
	go test ./...
