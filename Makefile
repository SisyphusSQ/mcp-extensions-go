.PHONY: fmt test vet build plugin

fmt:
	gofmt -w ui settings forms mentions resources internal examples/http examples/stdio examples/acceptance examples/agent-quickstart examples/form-mrtr

test:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/mcp-extensions-http ./examples/http
	go build -o bin/mcp-extensions-stdio ./examples/stdio
	go build -o bin/mcp-extensions-acceptance ./examples/acceptance
	go build -o bin/agent-quickstart ./examples/agent-quickstart
	go build -o bin/form-mrtr ./examples/form-mrtr

# Build a private local plugin; generated artifacts stay outside Git.
plugin:
	npm --prefix examples/frontend run build
	go build -o plugins/mcp-extensions-go/bin/mcp-extensions-stdio ./examples/stdio
	cp examples/frontend/dist/app.html plugins/mcp-extensions-go/assets/app.html
