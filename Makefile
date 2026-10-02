.PHONY: fmt test vet build plugin

fmt:
	gofmt -w ui settings forms mentions resources internal examples/http examples/stdio

test:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/mcp-extensions-http ./examples/http
	go build -o bin/mcp-extensions-stdio ./examples/stdio

# Build a private local plugin; generated artifacts stay outside Git.
plugin:
	npm --prefix examples/frontend run build
	go build -o plugins/mcp-extensions-go/bin/mcp-extensions-stdio ./examples/stdio
	cp examples/frontend/dist/app.html plugins/mcp-extensions-go/assets/app.html
