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

.PHONY: clean
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-*
	rm -rf dist/
	rm -rf completions/
