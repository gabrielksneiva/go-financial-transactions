.PHONY: test coverage cover-html lint fmt all coverage-html

WIN_DIR = /mnt/c/Users/gabri/Documents/

# Formata o código
fmt:
	go fmt ./...

# Roda o lint (requer o golangci-lint instalado)
lint:
	golangci-lint run

# Roda os testes com cobertura e gera arquivo coverage.out
coverage:
	go test ./... -coverprofile=coverage.out

# Gera coverage.html com gocov-html via go run
coverage-html: coverage
	go run github.com/axw/gocov/gocov@latest test ./... | \
	go run github.com/matm/gocov-html/cmd/gocov-html@latest > coverage.html
	cp coverage.html $(WIN_DIR)
	@echo "✔ coverage.html copiado para $(WIN_DIR)"

# Roda todos os testes com cobertura e exporta HTML
test:
	go test ./... -coverprofile=coverage.out
	go run github.com/axw/gocov/gocov@latest test ./... | \
	go run github.com/matm/gocov-html/cmd/gocov-html@latest > coverage.html
	cp coverage.html $(WIN_DIR)
	@echo "✔ Testes rodados e coverage.html copiado para $(WIN_DIR)"

# Roda tudo: fmt, test, lint
all: fmt test lint
