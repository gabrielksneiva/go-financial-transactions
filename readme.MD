
# 💸 Go Financial Transactions

> A performant and scalable financial transaction system built with Go, Kafka, PostgreSQL, and Redis, featuring secure blockchain withdrawals via TRON.

![Project Architecture](docs/assets/architecture.png)

[![Go](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![Tests](https://github.com/gabrielksneiva/go-financial-transactions/actions/workflows/go-ci.yml/badge.svg)](https://github.com/gabrielksneiva/go-financial-transactions/actions)
[![Coverage](https://img.shields.io/badge/coverage-89%25-brightgreen)](#)
[![License](https://img.shields.io/github/license/gabrielksneiva/go-financial-transactions)](LICENSE)

---

## 🧠 Project Overview

This repository presents a modern financial transaction system developed using the Go programming language. It is designed for high scalability, performance, and security—ideal for applications such as digital wallets, online banking platforms, and fintech APIs.

**Key features include:**

- 🔐 JWT-based authentication
- 🧵 Asynchronous transaction processing using Apache Kafka
- 🧠 Worker-based concurrency model in Go
- 📦 PostgreSQL with GORM for persistent storage
- 🚦 Redis for request rate limiting
- 🌐 Integration with the TRON blockchain for secure withdrawals
- ✅ >85% test coverage
- ⚙️ CI/CD via GitHub Actions

---

## 🧩 Architecture

The project follows a modular and layered architecture, ensuring code clarity, separation of concerns, and scalability.

```
        +------------------+
        |   API (Fiber)    |
        +--------+---------+
                 |
                 v
     +-----------+------------+
     |     Business Logic     |
     +-----------+------------+
                 |
     +-----------+------------+
     |      Repositories      |
     +-----+------------+-----+
           |            |
     +-----+         +--+-------------+
     | Kafka         | PostgreSQL     |
     +-----+         +----------------+
        |
   +----+-------+
   |  Workers   |
   +------------+
```

Kafka acts as a message broker between the API and background workers, allowing for asynchronous, distributed transaction processing.

---

## 🛠️ Technologies Used

- [Go 1.21](https://golang.org/)
- [Apache Kafka](https://kafka.apache.org/)
- [Fiber Web Framework](https://gofiber.io/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [GORM ORM](https://gorm.io/)
- [JWT](https://github.com/golang-jwt/jwt)
- [TRON Blockchain SDK](https://developers.tron.network/)
- [sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [GitHub Actions](https://github.com/features/actions)

---

## 🔐 Authentication

All secure routes require a valid JWT token. Authentication is handled via the `/api/login` endpoint, and token validation is enforced via middleware.

---

## 🧾 API Endpoints

| Method | Endpoint                     | Description                                | Authentication |
|--------|------------------------------|--------------------------------------------|----------------|
| POST   | `/api/register`              | Register a new user                        | ❌ No           |
| POST   | `/api/login`                 | Authenticate and receive JWT               | ❌ No           |
| POST   | `/api/deposit`               | Create a new deposit                       | ✅ Yes          |
| POST   | `/api/withdraw`              | Initiate a withdrawal via TRON blockchain  | ✅ Yes          |
| GET    | `/api/balance/:user_id`      | Retrieve user's current balance            | ✅ Yes          |
| GET    | `/api/statement/:user_id`    | Retrieve user's transaction statement      | ✅ Yes          |

> 🔄 Withdrawals are processed through the **TRON blockchain**, ensuring fast and secure crypto transfers.

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker + Docker Compose

### Installation

```bash
# Clone the repository
git clone https://github.com/gabrielksneiva/go-financial-transactions.git
cd go-financial-transactions

# Start services (PostgreSQL, Kafka, Redis)
docker-compose up -d

# Run the application
go run main.go
```

---

## 🧪 Running Tests

```bash
# Run all unit tests
go test ./... -coverprofile=coverage.out

# Check coverage
go tool cover -func=coverage.out
```

> ✅ Unit tests are available for all core layers: services, handlers, middleware, repositories, and workers.

---

## 📁 Project Structure

```bash
.
├── api/               # HTTP handlers and routing (Fiber)
├── config/            # Configuration logic
├── consumer/          # Kafka consumer
├── client/            # Application clients
├── producer/          # Kafka producer
├── domain/            # Entities and interfaces
├── services/          # Business logic
├── workers/           # Transaction workers
├── repositories/      # Database access layer
├── mocks/             # Mocks for testing
├── main.go            # Application entry point
├── go.mod / go.sum    # Dependencies
└── README.md          # Project documentation
```

---

## 🧠 Key Concepts

- Clean and modular architecture
- Kafka-based asynchronous event processing
- Goroutines and channel patterns
- Rate limiting with Redis
- Blockchain integration using the TRON SDK
- Extensive unit testing with mocking
- CI/CD pipelines with GitHub Actions

---

## 🔮 Future Enhancements

- [ ] Add Swagger/OpenAPI documentation
- [ ] Integrate with external custody services (e.g. Fireblocks)
- [ ] Enable observability with Prometheus and Grafana
- [ ] Deploy on Render, Railway or AWS ECS
- [ ] Implement audit logs for compliance

---

## 👨‍💻 Author

**Gabriel Kryzanovski dos Santos Neiva**  
[GitHub](https://github.com/gabrielksneiva) • [LinkedIn](https://linkedin.com/in/gabrielksneiva)

---
