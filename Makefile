SOURCE_FILES?=./...
TEST_PATTERN?=.
TEST_OPTIONS?=

# Install all the build and lint dependencies
setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
	curl -L https://git.io/misspell | sh
	go mod download
.PHONY: setup

# Run all the tests
test:
	go test $(TEST_OPTIONS) -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: test

# Run all the tests and opens the coverage report
cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

# gofmt all go files
fmt:
	find . -name '*.go' | while read -r file; do gofmt -w -s "$$file"; done
.PHONY: fmt

# Run all the linters
lint:
	# golangci-lint does not seem to currently support Go 1.13. See:
	# https://github.com/golangci/golangci-lint/issues/535
	#./bin/golangci-lint run ./...
	./bin/misspell -error **/*.go
.PHONY: lint

# Build a beta version of systools
build:
	go build -o systools cmd/systools/main.go
.PHONY: build

.DEFAULT_GOAL := build
