# Paisa

Paisa is being reframed as a loyalty-platform backend for companies that want to
offer customer rewards without building benefit management, member management,
reward calculation, and fulfillment infrastructure in house.

The current vertical slice is intentionally backend-first. It uses readable
`partner_key` URL scoping instead of auth while the core domain is still being
validated.

## Current Scope

- Partner creation and lookup by `partner_key`
- Partner programs and published rule versions
- Rule groups for `stack`, `max_of`, and `waterfall` earn behavior
- Member creation, member identifiers, accounts, and active program enrollment
- Partner-scoped transaction ingestion with idempotency
- Async transaction processing job
- Reward calculation from published rules only
- Refund/reversal calculation from the original transaction trace
- Insert-only ledger entries and balance snapshots
- Manual point adjustments
- Ledger-derived liability export summaries

Deferred for now:

- Auth, sessions, API keys, partner users, and internal users
- Polished product frontend
- Redemption/coupon fulfillment workflow
- Reservation timeout and earned-point expiration jobs
- Production migrations

## Architecture

```mermaid
flowchart LR
    Client["localhost curl client"] --> API["Go HTTP API"]
    API --> HTTP["adapters/httpapi"]
    HTTP --> Ports["ports"]
    Ports --> PG["adapters/postgres"]
    PG --> Domain["domain"]
    PG --> DB[("PostgreSQL")]
    Docker["postgres-docker compose"] --> DB
    Schema["db/schema.sql"] --> Docker
    Schema --> PG
```

## Repository Layout

```text
accts-api/
  server.go
  adapters/
    httpapi/
    postgres/
      internal/repository/
  application/
  domain/
  ports/
  api/
    requests.md
db/
  schema.sql
docs/
frontend/
  index.html
  src/
postgres-docker/
  docker-compose.yml
  setup.sh
```

`db/schema.sql` is the shared schema source. Docker mounts it into Postgres init,
and the Go API reads the same file during local startup. Override the path with
`PAISA_SCHEMA_PATH` if needed.

## Local Setup

Start Postgres:

```bash
cd postgres-docker
export PAISA_POSTGRES_PASSWORD="<local-dev-password>"
docker compose up -d
```

To reset local data and re-run the schema init:

```bash
cd postgres-docker
export PAISA_POSTGRES_PASSWORD="<local-dev-password>"
docker compose down -v
docker compose up -d
```

Run the API:

```bash
cd accts-api
export PAISA_POSTGRES_PASSWORD="<local-dev-password>"
PAISA_SCHEMA_PATH=../db/schema.sql go run server.go
```

The API listens on `http://localhost:8080`.

Run the simple UI:

```bash
cd frontend
npm install
npm run dev
```

Open the Vite localhost URL printed by `npm run dev`. The page is an API-backed
partner-admin console that calls the backend at `http://localhost:8080`. Login
with a readable `partner_key`; the UI will create that local test partner if it
does not exist. Use **Seed API demo suite** to create programs, published rules,
members, member add-ons, transactions, calculations, balances, and ledger rows in
Postgres, then inspect the refreshed dashboard, programs, rules, members, and
transactions views.

## Smoke Test Localhost Clients

These examples assume `curl` and `jq` are installed. Start Postgres and the API
first, then run the whole block from the repo root.

```bash
set -euo pipefail

API="http://localhost:8080"
PARTNER_KEY="acme-demo-$(date +%s)"
CUSTOMER_ID="cust-1001"

echo "1. Health"
curl -fsS "$API/health" | jq .

echo "2. Create partner"
PARTNER=$(curl -fsS -X POST "$API/v1/partners" \
  -H "Content-Type: application/json" \
  -d "{\"partnerKey\":\"$PARTNER_KEY\",\"name\":\"Acme Demo\"}")
echo "$PARTNER" | jq .

echo "3. Create program"
PROGRAM=$(curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/programs" \
  -H "Content-Type: application/json" \
  -d '{"name":"Gold Rewards","tierCode":"gold","priority":1}')
PROGRAM_ID=$(echo "$PROGRAM" | jq -r '.id')
echo "$PROGRAM" | jq .

echo "4. Create draft rule version"
RULE_VERSION=$(curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/programs/$PROGRAM_ID/rule-versions" \
  -H "Content-Type: application/json" \
  -d '{
    "earnBasis": "eligible",
    "ruleGroups": [
      {
        "name": "Everyday earn",
        "resolutionStrategy": "max_of",
        "priority": 1,
        "rules": [
          {
            "ruleKey": "base_earn",
            "name": "Base earn",
            "ruleType": "points_per_dollar",
            "priority": 1,
            "status": "active",
            "formulaConfig": {"pointsPerDollar": 1}
          },
          {
            "ruleKey": "grocery_bonus",
            "name": "Grocery bonus",
            "ruleType": "points_per_dollar",
            "priority": 2,
            "status": "active",
            "eligibilityConfig": {"categories": ["grocery"]},
            "formulaConfig": {"pointsPerDollar": 5},
            "limits": [
              {"scope": "member", "period": "calendar_month", "maxPoints": 300}
            ]
          }
        ]
      }
    ]
  }')
RULE_VERSION_ID=$(echo "$RULE_VERSION" | jq -r '.id')
echo "$RULE_VERSION" | jq .

echo "5. Publish rule version"
curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/programs/$PROGRAM_ID/rule-versions/$RULE_VERSION_ID/publish" | jq .

echo "6. Create enrolled member"
MEMBER=$(curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/members" \
  -H "Content-Type: application/json" \
  -d "{
    \"externalCustomerId\": \"$CUSTOMER_ID\",
    \"programId\": \"$PROGRAM_ID\",
    \"identifiers\": [{\"type\":\"email\",\"value\":\"member-1001@example.invalid\"}]
  }")
MEMBER_ID=$(echo "$MEMBER" | jq -r '.member.id')
echo "$MEMBER" | jq .

echo "7. Ingest purchase"
PURCHASE_BODY="{
  \"externalTransactionId\": \"txn-1001\",
  \"externalCustomerId\": \"$CUSTOMER_ID\",
  \"type\": \"purchase\",
  \"currency\": \"USD\",
  \"subtotalMinor\": 10000,
  \"taxMinor\": 600,
  \"totalMinor\": 10600,
  \"eligibleMinor\": 10000,
  \"occurredAt\": \"2026-07-24T12:00:00Z\",
  \"lineItems\": [
    {
      \"externalLineId\": \"line-1\",
      \"sku\": \"sku-grocery\",
      \"category\": \"grocery\",
      \"quantity\": 1,
      \"subtotalMinor\": 10000,
      \"taxMinor\": 600,
      \"totalMinor\": 10600,
      \"eligibleMinor\": 10000
    }
  ]
}"
PURCHASE=$(curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/ingest/transactions" \
  -H "Content-Type: application/json" \
  -d "$PURCHASE_BODY")
PURCHASE_ID=$(echo "$PURCHASE" | jq -r '.id')
echo "$PURCHASE" | jq .

echo "8. Retry same purchase; should return same transaction event"
curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/ingest/transactions" \
  -H "Content-Type: application/json" \
  -d "$PURCHASE_BODY" | jq .

echo "9. Same external transaction id with changed payload; should return 409"
CONFLICT_STATUS=$(curl -sS -o /tmp/paisa-conflict.json -w "%{http_code}" \
  -X POST "$API/v1/partners/$PARTNER_KEY/ingest/transactions" \
  -H "Content-Type: application/json" \
  -d "{
    \"externalTransactionId\": \"txn-1001\",
    \"externalCustomerId\": \"$CUSTOMER_ID\",
    \"type\": \"purchase\",
    \"currency\": \"USD\",
    \"eligibleMinor\": 20000
  }")
echo "status=$CONFLICT_STATUS"
cat /tmp/paisa-conflict.json | jq .

echo "10. Process accepted transaction events"
curl -fsS -X POST "$API/v1/jobs/process-transaction-events" | jq .

echo "11. Inspect purchase calculation, balance, and ledger"
curl -fsS "$API/v1/partners/$PARTNER_KEY/transactions/$PURCHASE_ID/calculation" | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/members/$MEMBER_ID/balance" | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/members/$MEMBER_ID/ledger" | jq .

echo "12. Ingest a half refund against the original purchase"
REFUND=$(curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/ingest/transactions" \
  -H "Content-Type: application/json" \
  -d "{
    \"externalTransactionId\": \"refund-1001\",
    \"externalCustomerId\": \"$CUSTOMER_ID\",
    \"originalExternalTransactionId\": \"txn-1001\",
    \"type\": \"refund\",
    \"currency\": \"USD\",
    \"subtotalMinor\": -5000,
    \"taxMinor\": -300,
    \"totalMinor\": -5300,
    \"eligibleMinor\": 5000,
    \"occurredAt\": \"2026-07-24T13:00:00Z\",
    \"lineItems\": [
      {
        \"externalLineId\": \"refund-line-1\",
        \"sku\": \"sku-grocery\",
        \"category\": \"grocery\",
        \"quantity\": 1,
        \"subtotalMinor\": -5000,
        \"taxMinor\": -300,
        \"totalMinor\": -5300,
        \"eligibleMinor\": 5000
      }
    ]
  }")
REFUND_ID=$(echo "$REFUND" | jq -r '.id')
echo "$REFUND" | jq .

echo "13. Process refund and inspect reversal"
curl -fsS -X POST "$API/v1/jobs/process-transaction-events" | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/transactions/$REFUND_ID/calculation" | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/members/$MEMBER_ID/balance" | jq .

echo "14. Create manual adjustment"
curl -fsS -X POST "$API/v1/partners/$PARTNER_KEY/members/$MEMBER_ID/adjustments" \
  -H "Content-Type: application/json" \
  -d '{"availableDelta":25,"reservedDelta":0,"expiredDelta":0,"reason":"smoke courtesy adjustment"}' | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/members/$MEMBER_ID/balance" | jq .

echo "15. Generate and list liability export"
curl -fsS -X POST "$API/v1/jobs/generate-ledger-liability-export" \
  -H "Content-Type: application/json" \
  -d "{\"partnerKey\":\"$PARTNER_KEY\",\"businessDate\":\"$(date +%F)\"}" | jq .
curl -fsS "$API/v1/partners/$PARTNER_KEY/exports/ledger-liability" | jq .
```

Expected high-level result:

- The purchase earns `300` points because the grocery `5x` rule beats base earn
  but is capped at `300`.
- The half refund reverses `150` points from the original calculation trace.
- The manual adjustment adds `25` points.
- Final available balance should be `175`.

## Developer Checks

```bash
cd accts-api
go test ./...
go build ./...
```

## Next Testing Step

The next quality jump should be DB-backed service tests that exercise real
Postgres transactions, row locking, and schema constraints. After that, add a
small HTTP smoke test script so the README flow can be run automatically.
