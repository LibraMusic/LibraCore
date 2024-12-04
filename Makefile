BINARY_NAME = libra

.PHONY: build
build:
	go build -v -o ${BINARY_NAME}

.PHONY: run
run:
	go run -v .

.PHONY: test
test:
	go test ./...

.PHONY: test_integration
test_integration:
	go test -tags=integration ./...

.PHONY: deps
deps:
	go mod download

.PHONY: clean
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	rm -rf dist/
	rm -rf completions/
