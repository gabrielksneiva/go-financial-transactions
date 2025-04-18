
# 💸 Financial Kafka Transaction Processor

This is a backend system designed to process financial transactions using Go, Kafka, and GORM. It receives deposit and withdrawal requests via an HTTP API, queues them in Kafka, and asynchronously processes them using a worker pool that updates the database.

---

## 🚀 Features

- Kafka Producer and Consumer integration
- Asynchronous processing with worker pool
- Persistent storage using PostgreSQL and GORM
- HTTP API built with Fiber
- Domain-driven structure with separation of concerns
- Balance and transaction statement queries

---

## 📦 Requirements

- Go 1.20+
- Kafka
- PostgreSQL

---

## ⚙️ Setup

1. **Clone the repository**

```bash
git clone https://github.com/yourusername/financial-kafka-processor.git
cd financial-kafka-processor
```

2. **Install dependencies**

```bash
go mod tidy
```

3. **Start PostgreSQL and Kafka**

You can use Docker or local services. Ensure Kafka is running on `localhost:9092` and PostgreSQL on `localhost:5432` with database name, user, and password configured in `repositories.InitDatabase()`.

4. **Run the application**

```bash
go run main.go
```

---

## 📘 API Endpoints

| Method | Endpoint               | Description               |
|--------|------------------------|---------------------------|
| POST   | `/deposit`             | Submit a deposit          |
| POST   | `/withdraw`            | Submit a withdrawal       |
| GET    | `/balance/:user_id`    | Get user balance          |
| GET    | `/statement/:user_id`  | Get transaction statement |

Example payload for deposit/withdraw:
```json
{
  "user_id": "123",
  "amount": 100.50
}
```

---

## 🧪 Load Testing

You can use [k6](https://k6.io/) to perform load tests:

1. Install k6.
2. Create a script (e.g. `loadtest.js`).
3. Run:
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
│
├── consumer/          # Kafka consumer logic
├── domain/            # Domain entities and interfaces
├── producer/          # Kafka producer logic
├── repositories/      # Repository implementations using GORM
├── services/          # Business logic (Deposit, Withdraw, Statement)
├── workers/           # Worker pool for background transaction processing
│
├── go.mod             # Go module file
├── go.sum             # Dependency checksum file
├── main.go            # Application entry point
└── README.md          # Project documentation
```

---

## 📄 License

MIT License © 2025 Gabriel Krzyzanovski dos Santos Neiva
