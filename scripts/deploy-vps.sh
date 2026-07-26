#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  echo "Usage: $0 <deploy-user> <server-ip> <repo-url> [branch]"
  exit 0
fi

if [[ $# -lt 3 ]]; then
  echo "Usage: $0 <deploy-user> <server-ip> <repo-url> [branch]"
  exit 1
fi

DEPLOY_USER="$1"
SERVER_IP="$2"
REPO_URL="$3"
BRANCH="${4:-main}"
APP_DIR="/opt/paisa"
SERVER_ENV_FILE="/etc/paisa/app.env"
LOCAL_APP_API_HOST_PORT="${APP_API_HOST_PORT:-8791}"
LOCAL_APP_WEB_HOST_PORT="${APP_WEB_HOST_PORT:-8790}"
LOCAL_PAISA_API_DOMAIN="${PAISA_API_DOMAIN:-${APP_API_DOMAIN:-}}"
LOCAL_PAISA_WEB_DOMAIN="${PAISA_WEB_DOMAIN:-${APP_DOMAIN:-}}"
LOCAL_PAISA_POSTGRES_DB="${PAISA_POSTGRES_DB:-project}"
LOCAL_PAISA_POSTGRES_USER="${PAISA_POSTGRES_USER:-project}"
LOCAL_PAISA_POSTGRES_PASSWORD="${PAISA_POSTGRES_PASSWORD:-}"
REPO_HOST=""

if [[ -z "$LOCAL_PAISA_API_DOMAIN" ]]; then
  echo "PAISA_API_DOMAIN is required." >&2
  exit 1
fi

if [[ -z "$LOCAL_PAISA_WEB_DOMAIN" ]]; then
  echo "PAISA_WEB_DOMAIN is required." >&2
  exit 1
fi

if [[ -z "$LOCAL_PAISA_POSTGRES_PASSWORD" ]]; then
  echo "PAISA_POSTGRES_PASSWORD is required." >&2
  exit 1
fi

if [[ "$REPO_URL" == git@*:* ]]; then
  REPO_HOST="${REPO_URL#git@}"
  REPO_HOST="${REPO_HOST%%:*}"
elif [[ "$REPO_URL" == http://* || "$REPO_URL" == https://* ]]; then
  REPO_HOST="${REPO_URL#*://}"
  REPO_HOST="${REPO_HOST%%/*}"
fi

ENV_PAYLOAD_BASE64="$(
  printf '%s\n' \
    "APP_API_HOST_PORT=${LOCAL_APP_API_HOST_PORT}" \
    "APP_WEB_HOST_PORT=${LOCAL_APP_WEB_HOST_PORT}" \
    "PAISA_API_DOMAIN=${LOCAL_PAISA_API_DOMAIN}" \
    "PAISA_WEB_DOMAIN=${LOCAL_PAISA_WEB_DOMAIN}" \
    "PAISA_POSTGRES_DB=${LOCAL_PAISA_POSTGRES_DB}" \
    "PAISA_POSTGRES_USER=${LOCAL_PAISA_POSTGRES_USER}" \
    "PAISA_POSTGRES_PASSWORD=${LOCAL_PAISA_POSTGRES_PASSWORD}" \
    "PAISA_HTTP_ADDR=0.0.0.0:8080" \
    "PAISA_SCHEMA_PATH=/app/db/schema.sql" \
    "PAISA_ALLOWED_ORIGINS=https://${LOCAL_PAISA_WEB_DOMAIN}" \
    | base64 | tr -d '\n'
)"

ssh "${DEPLOY_USER}@${SERVER_IP}" <<EOF
set -euo pipefail

if ! command -v docker >/dev/null 2>&1; then
  sudo apt-get update
  sudo apt-get install -y ca-certificates curl git
  sudo install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo tee /etc/apt/keyrings/docker.asc >/dev/null
  sudo chmod a+r /etc/apt/keyrings/docker.asc
  echo "deb [arch=\$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \$(. /etc/os-release && echo \$VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
  sudo apt-get update
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  sudo usermod -aG docker ${DEPLOY_USER}
fi

if [[ -n "${REPO_HOST}" ]]; then
  mkdir -p ~/.ssh
  chmod 700 ~/.ssh
  touch ~/.ssh/known_hosts
  chmod 600 ~/.ssh/known_hosts
  if ! ssh-keygen -F "${REPO_HOST}" >/dev/null 2>&1; then
    ssh-keyscan -H "${REPO_HOST}" >> ~/.ssh/known_hosts 2>/dev/null || true
  fi
fi

if [[ ! -d "${APP_DIR}/.git" ]]; then
  sudo mkdir -p "${APP_DIR}"
  sudo chown "${DEPLOY_USER}:${DEPLOY_USER}" "${APP_DIR}"
  if ! git clone --branch "${BRANCH}" "${REPO_URL}" "${APP_DIR}"; then
    echo "Git clone failed. If this repo is private, ensure server auth is configured."
    echo "Tip: use an HTTPS repo URL for public repos, or add deploy key/agent for SSH repos."
    exit 1
  fi
else
  cd "${APP_DIR}"
  git fetch origin
  git checkout "${BRANCH}"
  git pull --ff-only origin "${BRANCH}"
fi

cd "${APP_DIR}"

sudo mkdir -p "\$(dirname "${SERVER_ENV_FILE}")"
printf '%s' '${ENV_PAYLOAD_BASE64}' | base64 -d | sudo tee "${SERVER_ENV_FILE}" >/dev/null

sudo cp "${SERVER_ENV_FILE}" "${APP_DIR}/.env.prod"
sudo chown "${DEPLOY_USER}:${DEPLOY_USER}" "${APP_DIR}/.env.prod"

sudo docker compose --env-file "${APP_DIR}/.env.prod" -f "${APP_DIR}/docker-compose.prod.yml" up -d --build
sudo docker compose --env-file "${APP_DIR}/.env.prod" -f "${APP_DIR}/docker-compose.prod.yml" ps
EOF
