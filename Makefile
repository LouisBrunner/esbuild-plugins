# Args:
#  BUILD_TAGS: build tags to pass to the go build command
#  TEST_OUT_DIR: override the path of the test output directory
#  COV_OUT_DIR: override the path of the coverage output directory
#  COV_IN_FILE: override the path of the coverage input file

BUILD_TAGS ?=
TEST_OUT_DIR ?= .
COV_OUT_DIR ?= $(TEST_OUT_DIR)
COV_IN_FILE ?= $(TEST_OUT_DIR)/coverage.out

GOFLAGS = -tags=$(BUILD_TAGS)

all: vet test build
.PHONY: all

download: go.sum
	go mod download
.PHONY: download

mod-tidy:
	go mod tidy
.PHONY: mod-tidy

test: download generate
	go test $(GOFLAGS) -v ./...
.PHONY: test

test-ci: download generate
	go run $(GOFLAGS) gotest.tools/gotestsum --junitfile $(TEST_OUT_DIR)/report.xml -- -coverprofile=$(TEST_OUT_DIR)/coverage.out -v ./...
.PHONY: test-ci

coverage-ci: download generate
	go tool cover -html=$(COV_IN_FILE) -o $(COV_OUT_DIR)/coverage.html
	go run $(GOFLAGS) github.com/t-yuki/gocover-cobertura < $(COV_IN_FILE) > $(COV_OUT_DIR)/cobertura.xml
.PHONY: coverage-ci

vet: download generate
	gofmt -d -e -s .
	go vet $(GOFLAGS) ./...
	go run $(GOFLAGS) honnef.co/go/tools/cmd/staticcheck ./...
.PHONY: vet

generate: download
.PHONY: generate
