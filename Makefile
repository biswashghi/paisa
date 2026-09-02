SHELL := /bin/bash

PAISA_POSTGRES_PASSWORD ?= paisa-local-only
PAISA_INTERNAL_ADMIN_TOKEN ?= paisa-local-admin-token
PAISA_ALLOWED_ORIGINS ?= http://127.0.0.1:5174
PAISA_PUBLIC_API_URL ?= http://127.0.0.1:8081
export PAISA_POSTGRES_PASSWORD PAISA_INTERNAL_ADMIN_TOKEN PAISA_ALLOWED_ORIGINS PAISA_PUBLIC_API_URL

PAISA_STAGING_ALLOWED_ORIGINS ?= http://127.0.0.1:15174
PAISA_STAGING_PUBLIC_API_URL ?= http://127.0.0.1:18081

PAISA_ENV = env PAISA_POSTGRES_PASSWORD="$(PAISA_POSTGRES_PASSWORD)" PAISA_INTERNAL_ADMIN_TOKEN="$(PAISA_INTERNAL_ADMIN_TOKEN)"
LOCAL_COMPOSE = $(PAISA_ENV) PAISA_ALLOWED_ORIGINS="$(PAISA_ALLOWED_ORIGINS)" PAISA_PUBLIC_API_URL="$(PAISA_PUBLIC_API_URL)" docker compose -p paisa-local -f compose.yml -f compose.local.yml
STAGING_PROJECT ?= paisa-staging
STAGING_COMPOSE = $(PAISA_ENV) PAISA_ALLOWED_ORIGINS="$(PAISA_STAGING_ALLOWED_ORIGINS)" PAISA_PUBLIC_API_URL="$(PAISA_STAGING_PUBLIC_API_URL)" docker compose -p $(STAGING_PROJECT) -f compose.yml -f compose.staging.yml
PRODUCTION_COMPOSE = $(PAISA_ENV) PAISA_ALLOWED_ORIGINS="$(PAISA_ALLOWED_ORIGINS)" PAISA_PUBLIC_API_URL="$(PAISA_PUBLIC_API_URL)" docker compose -p paisa -f compose.yml -f compose.production.yml

.PHONY: local-up local-test local-down local-reset docker-build staging-test production-validate deployment-test

local-up:
	$(LOCAL_COMPOSE) up -d --build --wait

local-test:
	$(LOCAL_COMPOSE) up -d --build --wait
	cd accts-api && PAISA_POSTGRES_HOST=127.0.0.1 PAISA_POSTGRES_PORT=$${PAISA_LOCAL_DB_PORT:-5243} go test ./...
	curl --fail --silent --show-error http://127.0.0.1:$${PAISA_LOCAL_API_PORT:-8081}/health >/dev/null
	curl --fail --silent --show-error http://127.0.0.1:$${PAISA_LOCAL_WEB_PORT:-5174}/ >/dev/null

local-down:
	$(LOCAL_COMPOSE) down

local-reset:
	$(LOCAL_COMPOSE) down --volumes --remove-orphans

docker-build:
	$(STAGING_COMPOSE) build paisa-api paisa-web

staging-test:
	@set -uo pipefail; rm -f staging.log; \
	  cleanup() { $(STAGING_COMPOSE) down --volumes --remove-orphans; }; \
	  trap cleanup EXIT; \
	  status=0; \
	  $(STAGING_COMPOSE) up -d $${STAGING_NO_BUILD:+--no-build} --wait || status=$$?; \
	  if [[ $$status -eq 0 ]]; then curl --fail --silent --show-error http://127.0.0.1:$${PAISA_STAGING_API_PORT:-18081}/health >/dev/null || status=$$?; fi; \
	  if [[ $$status -eq 0 ]]; then curl --fail --silent --show-error http://127.0.0.1:$${PAISA_STAGING_WEB_PORT:-15174}/ >/dev/null || status=$$?; fi; \
	  if [[ $$status -eq 0 ]]; then curl --fail --silent --show-error http://127.0.0.1:$${PAISA_STAGING_WEB_PORT:-15174}/config.js | grep -F "$(PAISA_STAGING_PUBLIC_API_URL)" >/dev/null || status=$$?; fi; \
	  if [[ $$status -ne 0 ]]; then $(STAGING_COMPOSE) logs --no-color > staging.log 2>&1 || true; fi; \
	  exit $$status

production-validate:
	$(PRODUCTION_COMPOSE) config --quiet
	$(MAKE) deployment-test

deployment-test:
	bash tests/deploy-production.test.sh
