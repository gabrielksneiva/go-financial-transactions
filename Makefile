.PHONY: fmt lint test coverage coverage-html all ensure-tools

WIN_DIR = /mnt/c/Users/gabri/Documents/

ensure-tools:
	@command -v golangci-lint >/dev/null || ( \
		echo "📦 Instalando golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; )

fmt:
	@echo "🧹 Formatando código..."
	go fmt ./...

lint: ensure-tools
	@echo "🔍 Rodando linter..."
	$(shell go env GOPATH)/bin/golangci-lint run

test:
	@echo "🧪 Rodando testes..."
	go test ./... -v -coverprofile=coverage.out

coverage-html: test
	@echo "📊 Gerando coverage.html..."
	go run github.com/axw/gocov/gocov@latest test ./... | \
	go run github.com/matm/gocov-html/cmd/gocov-html@latest > coverage.html
	cp coverage.html $(WIN_DIR)
	@echo "✔ coverage.html copiado para $(WIN_DIR)"

all: fmt lint test
	@echo "✅ Tudo pronto!"
