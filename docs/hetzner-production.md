# Hetzner Production Runbook

Paisa production runs on the shared Hetzner VPS behind the shared Caddy reverse
proxy. The public web console and API use separate domains, while app and
database ports stay private.

## Required Deployment Environment

```bash
export PAISA_WEB_DOMAIN="paisa.example.com"
export PAISA_API_DOMAIN="api.paisa.example.com"
export PAISA_POSTGRES_PASSWORD="<production-password>"
export APP_WEB_HOST_PORT="8790"
export APP_API_HOST_PORT="8791"
```

`PAISA_POSTGRES_PASSWORD` is copied to `/etc/paisa/app.env` on the server and
then to `/opt/paisa/.env.prod` for Docker Compose.

## Deploy

Direct app deploy:

```bash
scripts/deploy-vps.sh <deploy-user> <server-ip> <repo-url> [branch]
```

Shared Terraform deploy:

```bash
cd /path/to/hetzner_tf
./scripts/deploy-vps-prod-from-tf.sh paisa main
```

The deploy script:
- installs Docker if needed
- clones or updates `/opt/paisa`
- writes `/etc/paisa/app.env`
- copies that env to `/opt/paisa/.env.prod`
- runs `docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build`

## Runtime

Production compose services:
- `paisa-web`: React console served by Caddy on localhost port `8790`
- `paisa-api`: Go API on localhost port `8791`
- `paisa-postgres`: internal Postgres with a named Docker volume

The API reads `/app/db/schema.sql` and applies the schema during startup.

## Verify

```bash
curl -f https://api.paisa.example.com/health
curl -I https://paisa.example.com
docker compose --env-file .env.prod -f docker-compose.prod.yml ps
docker compose --env-file .env.prod -f docker-compose.prod.yml logs --tail=100 paisa-api
```
