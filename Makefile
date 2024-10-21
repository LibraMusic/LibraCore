BINARY_NAME = libra

.PHONY: build
build:
	go build -v -o ${BINARY_NAME}

.PHONY: run
run:
	go run -v .

.PHONY: test
test:
 go test

.PHONY: test_coverage
test_coverage:
 go test -coverprofile=cover.out

.PHONY: dep
dep:
 go mod download

.PHONY: build_all
build_all:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux-amd64
	GOARCH=arm64 GOOS=linux go build -o ${BINARY_NAME}-linux-arm64
	GOARCH=amd64 GOOS=windows go build -o ${BINARY_NAME}-windows-amd64
	GOARCH=arm64 GOOS=windows go build -o ${BINARY_NAME}-windows-arm64
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin-amd64
	GOARCH=arm64 GOOS=darwin go build -o ${BINARY_NAME}-darwin-arm64

.PHONY: clean
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}-linux-amd64
	rm -f ${BINARY_NAME}-linux-arm64
	rm -f ${BINARY_NAME}-windows-amd64
	rm -f ${BINARY_NAME}-windows-arm64
	rm -f ${BINARY_NAME}-darwin-amd64
	rm -f ${BINARY_NAME}-darwin-arm64
