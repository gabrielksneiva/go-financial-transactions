
# 💸 Financial Kafka Consumer Project

Este é um projeto de exemplo que simula um sistema financeiro com operações de depósito, saque, saldo e extrato. A comunicação entre os serviços é feita via Apache Kafka, com persistência em PostgreSQL, API em Go + Fiber e uso de GORM como ORM.

## 📦 Tecnologias Utilizadas

- Go
- Fiber
- Kafka
- PostgreSQL
- GORM
- Docker

📁 Estrutura do Projeto

.
├── api/               # Handlers da API e configuração do Fiber
│   ├── api.go         # Inicialização do servidor HTTP
│   ├── handlers.go    # Funções HTTP (deposit, withdraw, balance, statement)
│   └── routes.go      # Definição das rotas da aplicação
│
├── consumer/          # Lógica do consumidor Kafka
│
├── domain/            # Definição das entidades e interfaces
│
├── producer/          # Lógica do produtor Kafka
│
├── repositories/      # Implementações dos repositórios com GORM
│
├── services/          # Regras de negócio (Depósito, Saque, Extrato)
│
├── workers/           # Worker pool para processar transações recebidas
│
├── go.mod             # Arquivo de definição de módulos Go
├── go.sum             # Checksum das dependências
├── main.go            # Ponto de entrada da aplicação
└── README.md          # Documentação do projeto

## 🚀 Como Rodar o Projeto

### Pré-requisitos

- Go 1.21+
- Docker & Docker Compose

### 1. Suba os containers (Kafka + PostgreSQL)

docker-compose up -d

### 2. Rode a aplicação Go

go run main.go

## 📬 Endpoints da API

| Método | Rota                     | Descrição                  |
|--------|--------------------------|----------------------------|
| POST   | /deposit                 | Envia depósito para Kafka  |
| POST   | /withdraw                | Envia saque para Kafka     |
| GET    | /balance/:user_id        | Retorna saldo atual        |
| GET    | /statement/:user_id      | Retorna extrato completo   |

### Exemplo de POST /deposit

{
  "user_id": "user123",
  "amount": 100.0
}

## 🛠 Teste de Carga (Opcional)

Você pode usar k6 para realizar testes de carga:

k6 run load-test.js

## 📊 Visualização do Kafka

Ferramentas recomendadas para visualizar mensagens Kafka:

- Kafdrop
- Kafka Tool

## 📌 Possíveis Melhorias Futuras

- Autenticação com JWT
- Integração com Fireblocks (saques)
- Métricas com Prometheus/Grafana
- Testes unitários e integração

## 👨‍💻 Autor

Desenvolvido com 💙 em Go para estudos de Kafka + arquitetura limpa.
