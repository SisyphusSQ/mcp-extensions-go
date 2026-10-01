.PHONY: fmt test vet build

fmt:
	gofmt -w ui examples/http

test:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/mcp-extensions-http ./examples/http
