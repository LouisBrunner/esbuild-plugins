all: vet test

download:
    go mod download

mod-tidy:
    go mod tidy

test *args: download
    go test {{ args }} -v ./...

test-ci output=".": download
    go tool gotestsum --junitfile {{ output }}/report.xml -- -coverprofile={{ output }}/coverage.out -v ./...

coverage-ci output=".": download
    go tool cover -html={{ output }}/coverage.out -o {{ output }}/coverage.html
    go tool gocover-cobertura < {{ output }}/coverage.out > {{ output }}/cobertura.xml

vet *args: download
    gofmt -d -e -s .
    go vet {{ args }} ./...
    go tool staticcheck ./...
