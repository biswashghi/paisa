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
export PAISA_INTERNAL_ADMIN_TOKEN="local-internal-admin-token"
```

Run checks:

```sh
go test ./...
go build ./...
```

Create a local partner admin before logging into the partner portal:

```sh
curl -X POST http://localhost:8080/internal/v1/partners/onboard \
  -H 'content-type: application/json' \
  -H 'X-Paisa-Internal-Admin-Token: local-internal-admin-token' \
  -d '{"partnerKey":"acme-retail","partnerName":"Acme Retail","adminEmail":"admin@acme-retail.test","adminName":"Acme Admin","adminPassword":"AcmeAdmin123"}'
```
