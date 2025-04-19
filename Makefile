.PHONY: test test-integration coverage coverage-integration cover-html lint fmt all coverage-html

WIN_DIR = /mnt/c/Users/gabri/Documents/

# Formata o código
fmt:
	go fmt ./...

# Roda o lint (requer o golangci-lint instalado)
lint:
	golangci-lint run

# Roda os testes com cobertura e gera coverage.out
coverage:
	go test ./... -coverprofile=coverage.out

# Roda os testes de integração com coverage
coverage-integration:
	go test -tags=integration ./integration -coverprofile=coverage-integration.out

# Gera coverage.html com gocov-html via go run
coverage-html: coverage
	go run github.com/axw/gocov/gocov@latest test ./... | \
	go run github.com/matm/gocov-html/cmd/gocov-html@latest > coverage.html
	cp coverage.html $(WIN_DIR)
	@echo "✔ coverage.html copiado para $(WIN_DIR)"

# Roda todos os testes + coverage + exporta HTML
test:
	go test ./... -coverprofile=coverage.out
	go run github.com/axw/gocov/gocov@latest test ./... | \
	go run github.com/matm/gocov-html/cmd/gocov-html@latest > coverage.html
	cp coverage.html $(WIN_DIR)
	@echo "✔ Testes rodados e coverage.html copiado para $(WIN_DIR)"

# Testes de integração com tag build
test-integration:
	go test -tags=integration ./integration -v -cover
	@echo "✔ Testes de integração finalizados com sucesso"

# Roda tudo: fmt, test, lint
all: fmt test lint
