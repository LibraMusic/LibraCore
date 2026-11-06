binary_name := "libra"
tags := env("BUILD_TAGS", "")

alias test := test-unit
alias check := lint

[default]
[group('build')]
build: deps gen
  go build -v -tags "{{ tags }}" -o {{ binary_name }} cmd/libra/main.go

[group('build')]
deps:
  go mod download

[group('build')]
gen:
  go generate ./...

[group('test')]
test-unit: deps
    @echo "Running unit tests..."
    go test ./...

[group('test')]
test-integration: deps
    @echo "Running integration tests..."
    go test -tags=integration ./...

[group('test')]
test-all: test-unit test-integration

[group('check')]
lint: lint-go lint-python lint-shell

[group('check')]
lint-go:
  {{ assert(shell('command -v golangci-lint || true') != "", "golangci-lint is not installed. You can install it from https://golangci-lint.run/welcome/install/") }}
  golangci-lint run --fix

[group('check')]
lint-python:
  {{ assert(shell('command -v ruff || true') != "", "ruff is not installed. You can install it from https://docs.astral.sh/ruff/installation/") }}
  ruff check --fix
  ruff format

[group('check')]
lint-shell:
  {{ assert(shell('command -v shellcheck || true') != "", "shellcheck is not installed. You can install it from https://github.com/koalaman/shellcheck#installing") }}
  find . -type f -name '*.sh' -exec shellcheck {} +

[group('misc')]
clean:
  rm -f -- "{{ binary_name }}"
  rm -f -- "{{ binary_name }}"-*
  rm -rf dist completions manpages .ruff_cache
  go clean
