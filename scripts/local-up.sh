#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESET_LOCAL_DATA="${PAISA_LOCAL_RESET:-0}"

for arg in "$@"; do
  case "${arg}" in
    --fresh|--reset)
      RESET_LOCAL_DATA="1"
      ;;
    --help|-h)
      cat <<EOF
Usage: ./scripts/local-up.sh [--fresh]

Starts local Postgres, API, and frontend.

Options:
  --fresh    Remove the local Postgres Docker volume before startup.

Environment overrides:
  PAISA_POSTGRES_PASSWORD
  PAISA_FRONTEND_PORT
  PAISA_LOCAL_ADMIN_EMAIL
  PAISA_LOCAL_ADMIN_PASSWORD
EOF
      exit 0
      ;;
    *)
      echo "Unknown option: ${arg}" >&2
      exit 1
      ;;
  esac
done

export PAISA_POSTGRES_PASSWORD="${PAISA_POSTGRES_PASSWORD:-local-dev-password}"
export PAISA_POSTGRES_HOST="${PAISA_POSTGRES_HOST:-localhost}"
export PAISA_POSTGRES_PORT="${PAISA_POSTGRES_PORT:-5243}"
export PAISA_POSTGRES_USER="${PAISA_POSTGRES_USER:-project}"
export PAISA_POSTGRES_DB="${PAISA_POSTGRES_DB:-project}"
export PAISA_INTERNAL_ADMIN_TOKEN="${PAISA_INTERNAL_ADMIN_TOKEN:-local-internal-admin-token}"
export PAISA_HTTP_ADDR="${PAISA_HTTP_ADDR:-127.0.0.1:8080}"
export PAISA_API_URL="${PAISA_API_URL:-http://127.0.0.1:8080}"
export PAISA_FRONTEND_HOST="${PAISA_FRONTEND_HOST:-127.0.0.1}"
export PAISA_FRONTEND_PORT="${PAISA_FRONTEND_PORT:-5173}"
export PAISA_FRONTEND_URL="${PAISA_FRONTEND_URL:-http://${PAISA_FRONTEND_HOST}:${PAISA_FRONTEND_PORT}}"
export PAISA_ALLOWED_ORIGINS="${PAISA_ALLOWED_ORIGINS:-http://localhost:${PAISA_FRONTEND_PORT},${PAISA_FRONTEND_URL}}"
export PAISA_LOCAL_PARTNER_KEY="${PAISA_LOCAL_PARTNER_KEY:-acme-retail}"
export PAISA_LOCAL_PARTNER_NAME="${PAISA_LOCAL_PARTNER_NAME:-Acme Retail}"
export PAISA_LOCAL_ADMIN_EMAIL="${PAISA_LOCAL_ADMIN_EMAIL:-admin@acme-retail.test}"
export PAISA_LOCAL_ADMIN_NAME="${PAISA_LOCAL_ADMIN_NAME:-Acme Admin}"
export PAISA_LOCAL_ADMIN_PASSWORD="${PAISA_LOCAL_ADMIN_PASSWORD:-AcmeAdmin123}"

API_PID=""
WEB_PID=""

cleanup() {
  if [[ -n "${WEB_PID}" ]] && kill -0 "${WEB_PID}" 2>/dev/null; then
    kill "${WEB_PID}" 2>/dev/null || true
  fi
  if [[ -n "${API_PID}" ]] && kill -0 "${API_PID}" 2>/dev/null; then
    kill "${API_PID}" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

for command_name in docker go npm curl; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "${command_name} is required." >&2
    exit 1
  fi
done

echo "Starting Postgres on localhost:${PAISA_POSTGRES_PORT}..."
(
  cd "${ROOT_DIR}/postgres-docker"
  if [[ "${RESET_LOCAL_DATA}" == "1" ]]; then
    echo "Resetting local Postgres data volume..."
    docker compose down -v
  fi
  docker compose up -d
  for attempt in {1..60}; do
    if docker compose exec -T postgres pg_isready -U "${PAISA_POSTGRES_USER}" -d "${PAISA_POSTGRES_DB}" >/dev/null 2>&1; then
      exit 0
    fi
    sleep 1
  done
  echo "Postgres did not become ready." >&2
  exit 1
)

echo "Starting API on ${PAISA_API_URL}..."
(
  cd "${ROOT_DIR}/accts-api"
  PAISA_SCHEMA_PATH="${ROOT_DIR}/db/schema.sql" go run server.go
) &
API_PID="$!"

for attempt in {1..30}; do
  if curl --fail --silent "${PAISA_API_URL}/health" >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "${API_PID}" 2>/dev/null; then
    echo "API exited before becoming healthy." >&2
    wait "${API_PID}" || true
    exit 1
  fi
  sleep 1
done

if ! curl --fail --silent "${PAISA_API_URL}/health" >/dev/null 2>&1; then
  echo "API did not become healthy at ${PAISA_API_URL}/health." >&2
  exit 1
fi

echo "Creating local partner admin ${PAISA_LOCAL_ADMIN_EMAIL}..."
ONBOARD_PAYLOAD="$(
  printf '{"partnerKey":"%s","partnerName":"%s","adminEmail":"%s","adminName":"%s","adminPassword":"%s"}' \
    "${PAISA_LOCAL_PARTNER_KEY}" \
    "${PAISA_LOCAL_PARTNER_NAME}" \
    "${PAISA_LOCAL_ADMIN_EMAIL}" \
    "${PAISA_LOCAL_ADMIN_NAME}" \
    "${PAISA_LOCAL_ADMIN_PASSWORD}"
)"
ONBOARDED="0"
for attempt in {1..30}; do
  if curl --fail --silent --show-error -X POST "${PAISA_API_URL}/internal/v1/partners/onboard" \
    -H 'content-type: application/json' \
    -H "X-Paisa-Internal-Admin-Token: ${PAISA_INTERNAL_ADMIN_TOKEN}" \
    -d "${ONBOARD_PAYLOAD}" >/dev/null; then
    ONBOARDED="1"
    break
  fi
  if ! kill -0 "${API_PID}" 2>/dev/null; then
    echo "API exited during local partner onboarding." >&2
    wait "${API_PID}" || true
    exit 1
  fi
  sleep 1
done

if [[ "${ONBOARDED}" != "1" ]]; then
  echo "Could not create local partner admin after retries." >&2
  exit 1
fi

echo "Starting frontend on ${PAISA_FRONTEND_URL}..."
(
  cd "${ROOT_DIR}/frontend"
  VITE_PAISA_API_URL="${PAISA_API_URL}" npm run dev -- --host "${PAISA_FRONTEND_HOST}" --port "${PAISA_FRONTEND_PORT}"
) &
WEB_PID="$!"

cat <<EOF

Paisa local services are running.

Frontend: ${PAISA_FRONTEND_URL}
API:      ${PAISA_API_URL}
Postgres: localhost:${PAISA_POSTGRES_PORT}
Fresh DB: ${RESET_LOCAL_DATA}

Default local login:
Email:    ${PAISA_LOCAL_ADMIN_EMAIL}
Password: ${PAISA_LOCAL_ADMIN_PASSWORD}

Partner onboarding is run automatically on startup. Override the default login
with PAISA_LOCAL_ADMIN_EMAIL and PAISA_LOCAL_ADMIN_PASSWORD.

Press Ctrl-C to stop the API and frontend. Postgres will keep running in Docker.
EOF

wait "${WEB_PID}"
