# Default recipe
default: lint test

fomat:
    go fmt ./...
lint:
    golangci-lint run

lint-fix:
    golangci-lint run --fix

test:
    go test -v -race ./...
