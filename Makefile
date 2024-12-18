BINARY_NAME = libra

.PHONY: build
build:
	go build -v -o ${BINARY_NAME}

.PHONY: test
test:
	go test ./...

.PHONY: test_integration
test_integration:
	go test -tags=integration ./...

# Available engines: sqlite, postgresql
.PHONY: create_migration
create_migration:
	migrate create -ext sql -dir db/migrations/$(engine) -seq $(name)

.PHONY: clean
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	rm -rf dist/
	rm -rf completions/
	rm -rf manpages/
