---
marp: true
title: Paisa Codebase Documentation
author: GitHub Copilot
date: 2025-06-16
---

# Paisa Codebase Documentation

---

## Overview

Paisa is a full-stack application for managing user accounts, wallet balances, and transactions. It consists of a Go backend API, a PostgreSQL database (with Docker support), and a modern HTML/JS/CSS frontend.

---

## Project Structure

```
accts-api/         # Go backend API
frontend/          # HTML, CSS, JS frontend
postgres-docker/   # Dockerized PostgreSQL setup
```

---

## Backend (Go API)

- **Location:** `accts-api/`
- **Key files:**
  - `server.go`: Main server setup, HTTP server, and routing
  - `api/api.go`: API endpoint logic (register, login, account, transactions)
  - `api/auth.go`: Auth middleware, login logic, session validation
  - `api/models.go`: Data models for requests/responses (Go structs)
- **Architecture:**
  - Uses [Gorilla Mux](https://github.com/gorilla/mux) for routing
  - Follows RESTful API design
  - Middleware for session authentication on all `/api/*` endpoints
  - Passwords are hashed with bcrypt before storage
  - Session tokens are generated and validated for each user session
  - Database access via `database/sql` and context-aware queries
- **Endpoints:**
  - `POST /register` — Register a new user (creates user and account)
  - `POST /login` — Login, returns session token
  - `GET /api/{accountId}/details` — Get account details
  - `PUT /api/update/balance/{accountId}` — Add transaction (with description, merchant code, balance)
  - `GET /api/list/transactions/{accountId}` — List transactions (paginated)
- **Session Management:**
  - On successful login, a session token (UUID) is generated and stored in `paisa.sessions`
  - The session token must be sent in the `Authorization` header for all `/api/*` requests
  - Middleware checks the session token and injects the user's accountId into the request context
- **Password Security:**
  - Registration: Passwords are hashed using bcrypt before being stored in `paisa.users`
  - Login: Passwords are checked using bcrypt's `CompareHashAndPassword`
  - Raw passwords are never stored or returned
- **Error Handling:**
  - All endpoints return JSON error messages with appropriate HTTP status codes
  - Common errors: invalid credentials, unauthorized, not found, bad request
- **Data Models:**
  - `User`: accountId, username, passwordHash
  - `Account`: accountId, balance
  - `Transaction`: id, accountId, description, merchantCode, balance, timestamp
  - `Session`: sessionId, accountId, createdAt
- **Pagination:**
  - Transaction listing supports pagination via query parameters (e.g., `?page=1&page_size=10`)
  - Returns a list of transactions for the requested page
- **Extensibility:**
  - Easy to add new endpoints by following the existing handler pattern in `api.go`
  - Models and database schema can be extended for new features
- **Testing:**
  - Endpoints can be tested with curl or Postman
  - Example curl requests are provided in the docs
- **Best Practices:**
  - Uses context for DB queries
  - Separates concerns: routing, models, handlers, middleware
  - Follows Go idioms for error handling and struct usage

---

## Database (PostgreSQL)

- **Location:** `postgres-docker/`
- **Key files:**
  - `psql.sql`: Schema for all tables
  - `docker-compose.yml`: Docker setup for Postgres
  - `setup.sh`: Script to start and initialize DB
- **Main tables:**
  - `paisa.users`: Users (accountid, username, password hash)
  - `paisa.accounts`: Wallet accounts (accountid, balance)
  - `paisa.transactions`: Transactions (id, accountid, description, merchant_code, balance, timestamp)
  - `paisa.sessions`: Session tokens

---

## Frontend

- **Location:** `frontend/`
- **Key files:**
  - `index.html`: Main dashboard (wallet, transactions)
  - `login.html`: Login page
  - `register.html`: Registration page
  - `static/js/`: JS logic (auth, account info, main)
  - `static/css/main.css`: Styling
- **Features:**
  - Responsive design with Bootstrap
  - Dynamic loading of wallet and transactions
  - Pagination for transactions
  - Session management via localStorage

---

## API Usage Examples

### Register
```sh
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "pass"}'
```

### Login
```sh
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "password": "pass"}'
```

### Add Transaction
```sh
curl -X PUT http://localhost:8080/api/update/balance/{accountId} \
  -H "Authorization: {sessionId}" \
  -H "Content-Type: application/json" \
  -d '{"balance": 100.00, "description": "Test", "merchantCode": "ABC123"}'
```

---

## Security Practices

- Passwords are never stored or transmitted in plaintext
- All passwords are hashed with bcrypt before storage
- Session tokens are required for all sensitive API calls
- CORS and CSRF protections are recommended for production

---

## Development & Setup

1. **Start Postgres:**
   ```sh
   cd postgres-docker
   ./setup.sh
   ```
2. **Run Backend:**
   ```sh
   cd accts-api
   go run server.go
   ```
3. **Run Frontend:**
   ```sh
   cd frontend
   python3 -m http.server 8081
   # or use any static file server
   ```

---

## Contributing

- Follow Go and JS best practices
- Keep API and DB schemas in sync
- Write clear commit messages
- Document new endpoints and features

---

## License

MIT License

---

## Contact

For questions or contributions, contact the maintainer via GitHub.
