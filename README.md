# Users Service (Go Backend)

Backend service sederhana untuk manajemen **Users** yang dibangun dengan Go, Gin, dan GORM. Project ini dibuat sebagai studi kasus praktik backend profesional dengan menerapkan clean architecture, error handling konsisten, testing, dan CI/CD.

## ✨ Fitur

- **CRUD Users API**: Create, List, Get, Update, Delete
- **Konsistensi error** via `apperr` + `httpx.ErrorMiddleware`
- **Audit logging** hook dengan zap logger
- **Unique index** pada `users.email`
- **Clean Architecture**: Handler → Service → Repository (GORM)
- **Unit & Integration Tests** (SQLite in-memory), coverage **≥70%**
- **GitHub Actions CI**: automated testing + coverage gate

## 📦 Tech Stack

- **Go** 1.22
- **Gin** (HTTP router)
- **GORM** (ORM)
- **SQLite** (testing), **PostgreSQL** (production)
- **Zap** (structured logger)

## 🗂 Struktur Project

```
.
├── cmd/
│   └── api/                    # Application entrypoint
├── internal/
│   ├── apperr/                # Typed application errors
│   ├── httpx/                 # HTTP response, middleware, audit
│   ├── logger/                # Global zap logger
│   └── users/                 # User domain (DTO, handler, service, repository, tests)
├── .github/workflows/
│   └── ci.yml                 # GitHub Actions CI/CD
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🚀 Quick Start

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Run Tests

```bash
# Using Makefile
make test

# Or directly with go
go test ./internal/users -cover -v
```

### 3. Start Server

```bash
go run ./cmd/api
```

Server akan berjalan di `http://localhost:8080`

### 4. Database Configuration

- **Development/Testing**: SQLite in-memory (otomatis di tests)
- **Production**: PostgreSQL dengan DSN

Untuk production, pastikan unique index sudah dibuat:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS uix_users_email ON users(email);
```

## 🔌 API Endpoints

Base URL: `http://localhost:8080/v1`

| Method | Endpoint        | Description    |
|--------|----------------|----------------|
| POST   | `/users`       | Create user    |
| GET    | `/users`       | List all users |
| GET    | `/users/:id`   | Get user by ID |
| PUT    | `/users/:id`   | Update user    |
| DELETE | `/users/:id`   | Delete user    |

### Request/Response Examples

#### Create User

```bash
POST /v1/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

#### Success Response

```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-09-08T12:00:00Z",
    "updated_at": "2025-09-08T12:00:00Z"
  }
}
```

#### Error Response

```json
{
  "error": "Bad Request",
  "message": "invalid request body",
  "code": 400,
  "time": "2025-09-08T12:00:00Z"
}
```

## 🧪 Testing & Coverage

### Run Tests

```bash
# Run tests with coverage
go test ./internal/users -coverprofile=cover.out

# View coverage summary
go tool cover -func=cover.out

# View coverage in browser
go tool cover -html=cover.out
```

### Coverage Requirements

- **Minimum coverage**: 70%
- CI akan gagal jika coverage di bawah threshold
- Tests menggunakan SQLite in-memory untuk isolation

## 🛠 Development

### Makefile Commands

```bash
make test          # Run tests
make test-cover    # Run tests with coverage report
make build         # Build binary
make run           # Run development server
make clean         # Clean build artifacts
```

### Environment Variables

```bash
# Database
DB_DSN=postgres://user:pass@localhost/dbname

# Server
PORT=8080
GIN_MODE=release

# Logging
LOG_LEVEL=info
```

## 🔄 CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/ci.yml`) melakukan:

1. **Setup Go** environment
2. **Install dependencies**
3. **Run tests** dengan coverage
4. **Coverage gate** - fail jika < 70%
5. **Build** aplikasi

## 🛡 Branch Strategy

- **`main`**: Production-ready code, protected branch (PR + CI required)
- **`develop`**: Integration branch untuk development
- **`feature/*`**: Feature branches untuk development baru

### Workflow

1. Create feature branch dari `develop`
2. Implement changes dengan tests
3. Create PR ke `develop`
4. Setelah review, merge ke `develop`
5. Create release PR dari `develop` ke `main`

## 🏗 Architecture

Project mengikuti **Clean Architecture** principles:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Handler   │───▶│   Service   │───▶│ Repository  │
└─────────────┘    └─────────────┘    └─────────────┘
      │                    │                   │
      ▼                    ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   HTTP/Gin  │    │ Business    │    │   GORM/DB   │
│             │    │   Logic     │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Layer Responsibilities

- **Handler**: HTTP request/response, validation, routing
- **Service**: Business logic, orchestration
- **Repository**: Data access, database operations

## 📝 Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add some amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Code Style

- Follow Go conventions
- Run `gofmt` before commit
- Add tests untuk new features
- Maintain minimum 70% coverage

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Support

Jika ada pertanyaan atau issues:

1. Check existing [Issues](https://github.com/yourusername/users-service/issues)
2. Create new issue dengan detail yang cukup
3. Atau contact maintainer

---

**Happy Coding! 🚀**
