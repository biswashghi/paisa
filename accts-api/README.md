# Loyalty API

Local configuration is environment-driven so secrets are not committed.

Required for local Postgres:

```sh
export PAISA_POSTGRES_PASSWORD="<local-dev-password>"
```

Optional overrides:

```sh
export PAISA_POSTGRES_HOST="localhost"
export PAISA_POSTGRES_PORT="5243"
export PAISA_POSTGRES_USER="project"
export PAISA_POSTGRES_DB="project"
export PAISA_ALLOWED_ORIGINS="http://localhost:5173"
```

Run checks:

```sh
go test ./...
go build ./...
```
