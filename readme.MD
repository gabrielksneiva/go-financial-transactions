
# 💸 Financial Kafka Transaction Processor

This backend system is designed to process financial transactions using Go, Kafka, and GORM. It handles deposit and withdrawal requests via an HTTP API, queues them in Kafka, and processes them asynchronously using a worker pool that updates the database.

---

## 🚀 Features

- Seamless Kafka Producer and Consumer integration
- Asynchronous processing powered by a worker pool
- Persistent storage with PostgreSQL and GORM
- HTTP API built using Fiber
- Domain-driven design with clear separation of concerns
- Balance inquiries and transaction statement retrieval
- Transaction processing with Kafka for asynchronous operations
- Background transaction processing via a worker pool

---

## 📦 Requirements

- Go 1.20+
- Kafka
- PostgreSQL

---

## ⚙️ Setup

1. **Clone the repository**

```bash
git clone https://github.com/gabrielksneiva/go-financial-transactions.git
cd go-financial-transactions
```

2. **Install dependencies**

```bash
go mod tidy
```

3. **Start PostgreSQL and Kafka**

You can use Docker or local services. Ensure Kafka is running on `localhost:9092` and PostgreSQL on `localhost:5432` with the database name, user, and password configured in `repositories.InitDatabase()`.

4. **Run the application**

```bash
go run main.go
```

---

## 📘 API Endpoints

| Method | Endpoint               | Description               |
|--------|------------------------|---------------------------|
| POST   | `/deposit`             | Submit a deposit request  |
| POST   | `/withdraw`            | Submit a withdrawal request |
| POST   | `/user`                | Create a new user         |
| GET    | `/balance/:user_id`    | Retrieve user balance     |
| GET    | `/statement/:user_id`  | Retrieve transaction statement |

Example payload for deposit/withdraw requests:
```json
{
  "user_id": 123,
  "amount": 100.50
}
```

---

## 🧪 Load Testing

You can use [k6](https://k6.io/) to perform load testing:

1. Install k6.
2. Create a script (e.g., `loadtest.js`).
3. Run the test:
```bash
k6 run loadtest.js
```

---

## 🛠️ Project Structure

```text
.
├── api/               # HTTP handlers and Fiber setup
│   ├── api.go         # Initializes the HTTP server
│   ├── handlers.go    # Request handlers
│   └── routes.go      # Route definitions
│── config/            # Application configuration
├── consumer/          # Kafka consumer logic
├── domain/            # Domain entities and interfaces
├── producer/          # Kafka producer logic
├── repositories/      # Repository implementations using GORM
├── services/          # Business logic (Deposit, Withdraw, Statement)
├── workers/           # Worker pool for background transaction processing
│
├── docker-compose.yaml# Docker Compose configuration
├── go.mod             # Go module file
├── go.sum             # Dependency checksum file
├── main_test.go       # Application entry point tests
├── main.go            # Application entry point
├── Makefile           # Makefile for project automation
└── README.md          # Project documentation
```

---

## 📄 License

MIT License © 2023 Gabriel Krzyzanovski dos Santos Neiva

