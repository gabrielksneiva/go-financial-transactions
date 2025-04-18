.PHONY: test coverage cover-html lint fmt all coverage-html

# Formata o código
fmt:
	go fmt ./...

# Roda o lint (se tiver o golangci-lint instalado)
lint:
	golangci-lint run

# Gera um arquivo de cobertura
coverage:
	go test ./... -coverprofile=coverage.out

# Gera o HTML da cobertura e copia pro Windows
coverage-html: coverage
	go install github.com/axw/gocov/gocov@latest
	go install github.com/matm/gocov-html/cmd/gocov-html@latest
	gocov test ./... | gocov-html > coverage.html
	cp coverage.html /mnt/c/Users/gabri/Documents/
	@echo "✔ coverage.html copiado para /mnt/c/Users/gabri/Documents/"

# Roda todos os testes com cobertura + gera HTML
test:
	go test ./... -coverprofile=coverage.out
	go install github.com/axw/gocov/gocov@latest
	go install github.com/matm/gocov-html/cmd/gocov-html@latest
	gocov test ./... | gocov-html > coverage.html
	cp coverage.html /mnt/c/Users/gabri/Documents/
	@echo "✔ Testes rodados e coverage.html copiado para /mnt/c/Users/gabri/Documents/"

# Roda tudo
all: fmt test lint
