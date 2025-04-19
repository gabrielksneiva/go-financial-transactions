# 💸 Go Financial Transactions

> Asynchronous and scalable financial transaction system built with Go, Kafka, PostgreSQL, and Redis.

![Project Architecture](docs/assets/architecture.png)

[![Go](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org/)
[![Tests](https://github.com/gabrielksneiva/go-financial-transactions/actions/workflows/go-ci.yml/badge.svg)](https://github.com/gabrielksneiva/go-financial-transactions/actions)
[![Coverage](https://img.shields.io/badge/coverage-89%25-brightgreen)](#)
[![License](https://img.shields.io/github/license/gabrielksneiva/go-financial-transactions)](LICENSE)

---

## 🧠 About the Project

This project simulates a financial transaction system designed to be highly performant, asynchronous, and scalable — ideal for banking applications, digital wallets, and payment processors. It demonstrates:

- Asynchronous messaging via **Apache Kafka**
- Concurrent transaction processing with **Go workers**
- Request rate limiting with **Redis**
- Data consistency and storage with **PostgreSQL + GORM**
- Authentication using **JWT**
- Test coverage > 85%
- CI/CD using **GitHub Actions**

---

## 🧩 Architecture

The system is built using a clean and modular architecture.

```
        +---------------+
        |  API (Fiber)  |
        +------+--------+
               |
               v
     +---------+----------+
     | Services (Business)|
     +---------+----------+
               |
     +---------+----------+
     |    Repositories    |
     +----+----------+----+
          |          |
    +-----+      +---+---------+
    |Kafka|      | PostgreSQL  |
    +-----+      +-------------+
```

Kafka acts as a middleware between the HTTP layer and the processing workers. Workers consume messages asynchronously and update balances in the database.

---

## 🛠️ Technologies

- [Go 1.21](https://golang.org/)
- [Apache Kafka](https://kafka.apache.org/)
- [Fiber](https://gofiber.io/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [GORM](https://gorm.io/)
- [JWT](https://github.com/golang-jwt/jwt)
- [sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [GitHub Actions](https://github.com/features/actions)

---

## 🚀 Getting Started

### Prerequisites

- Docker + Docker Compose
- Go 1.21+

```bash
# Clone the project
git clone https://github.com/gabrielksneiva/go-financial-transactions.git
cd go-financial-transactions

# Start required services (Kafka, Redis, PostgreSQL)
docker-compose up -d

# Run the application
go run main.go
```

---

## 🧪 Tests

```bash
# Run all unit tests
go test ./... -coverprofile=coverage.out

# View coverage report
go tool cover -func=coverage.out
```

✅ The project includes comprehensive unit tests for services, handlers, middleware, workers, repositories, and Kafka components.

---

## 📁 Project Structure

```bash
.
├── api/               # HTTP handlers and routing (Fiber)
├── config/            # Configuration logic
├── consumer/          # Kafka consumer
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

- Goroutines, channels and context cancellation
- Dependency injection
- Testing with `sqlmock` and `testify`
- Message-driven architecture with Kafka
- CI/CD with GitHub Actions
- Redis-based rate limiting
- Clean code and modularity

---

## 🎯 Roadmap

- [ ] Swagger documentation for API endpoints
- [ ] Fireblocks API integration
- [ ] Automated deployment with Docker + Render or Railway
- [ ] Observability with Prometheus and Grafana

---

## 👨‍💻 Author

**Gabriel Kryzanovski dos Santos Neiva**  
[GitHub](https://github.com/gabrielksneiva) • [LinkedIn](https://linkedin.com/in/gabrielksneiva)
