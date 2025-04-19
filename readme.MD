
# 💸 Go Financial Transactions

[![Go](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![Tests](https://github.com/gabrielksneiva/go-financial-transactions/actions/workflows/go-ci.yml/badge.svg)](https://github.com/gabrielksneiva/go-financial-transactions/actions)
[![Coverage](https://img.shields.io/badge/Coverage-90%25-brightgreen)](https://github.com/gabrielksneiva/go-financial-transactions)
[![Go Report Card](https://goreportcard.com/badge/github.com/gabrielksneiva/go-financial-transactions)](https://goreportcard.com/report/github.com/gabrielksneiva/go-financial-transactions)

> Projeto backend financeiro em Go com arquitetura moderna, paralelismo e Kafka. Ideal para processar e auditar transações de usuários em tempo real.

---

## ✨ Funcionalidades

- 📤 Envio de depósitos e saques com controle de saldo
- 🧾 Extrato completo com transações e saldo atual
- ✅ Validação por JWT
- 📈 Limite de requisições via Redis (Rate Limiting)
- 🧵 Workers concorrentes processando mensagens Kafka
- 🧪 Alta cobertura de testes unitários (>90%)

---

## 📦 Stack

- **Go** 1.21
- **Fiber** (HTTP server)
- **Kafka** (mensageria)
- **PostgreSQL** + GORM
- **Redis** (rate limit)
- **GitHub Actions** (CI/CD)
- **sqlmock**, **testify** (testes)

---

## 🚀 Rodando o projeto

```bash
# Clone o repositório
git clone https://github.com/gabrielksneiva/go-financial-transactions.git
cd go-financial-transactions

# Instale dependências
go mod tidy

# Copie o .env de exemplo
cp .env.example .env

# Rode a aplicação
go run main.go
```

---

## 🧪 Rodando os testes

```bash
# Testes com coverage
make test

# Cobertura em HTML
make coverage-html
```

---

## 📂 Estrutura

```bash
.
├── api/            # Endpoints HTTP
├── consumer/       # Leitor Kafka
├── producer/       # Escritor Kafka
├── services/       # Regras de negócio
├── repositories/   # Acesso a dados
├── workers/        # Workers concorrentes
├── config/         # Config e setup da aplicação
├── mocks/          # Mocks de interfaces
├── main.go         # Entrada principal
```

---

## 🧠 Arquitetura

> Arquitetura hexagonal com serviços desacoplados e workers concorrentes para processamento paralelo de transações financeiras.

---

## ✅ Testes e Qualidade

- [x] Cobertura >90%
- [x] Testes unitários com mocks (`testify`, `sqlmock`)
- [x] Linters via `golangci-lint`
- [x] Pipeline CI GitHub Actions

---

## 🛣️ Roadmap

- [ ] Deploy com Docker + docker-compose
- [ ] Integração com Fireblocks API
- [ ] Dashboard admin para visualizar extratos
- [ ] Suporte a múltiplas moedas

---

## 👨‍💻 Autor

Desenvolvido por **Gabriel K. S. Neiva**

[LinkedIn](https://www.linkedin.com/in/gabrielneiva/)  
[Email](mailto:gabrielksneivagi@gmail.com)

---

## 📄 Licença

MIT
