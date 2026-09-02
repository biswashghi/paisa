# Shared VPS Production Runbook

Paisa production runs on a shared VPS behind the platform-managed Caddy reverse
proxy. The VPS may be hosted by Hetzner, OVHcloud, or another provider. The
public web console and API use separate domains, while app and database ports
stay private inside Docker.

## Required Deployment Environment

```bash
export PAISA_WEB_DOMAIN="paisa.example.com"
export PAISA_API_DOMAIN="api.paisa.example.com"
export PAISA_POSTGRES_PASSWORD="<production-password>"
export PAISA_INTERNAL_ADMIN_TOKEN="<production-admin-token>"
export PAISA_API_IMAGE="ghcr.io/owner/repo/paisa-api@sha256:..."
export PAISA_WEB_IMAGE="ghcr.io/owner/repo/paisa-web@sha256:..."
```

Bootstrap the VPS platform once before deploying the app. That creates the
external `vps-edge` network, starts shared Caddy, and installs the `vps-route`
and `vps-platform-check` helpers. `PAISA_POSTGRES_PASSWORD` and the two domains
are stored in root-owned `/etc/paisa/app.env` on the server.

## Deploy

Direct app deploy:

```bash
scripts/deploy-vps.sh <deploy-user> <server-ip>
```

That entrypoint only validates inputs and uploads the release bundle. The
readable VPS-side sequence is `scripts/remote/deploy-production.sh`, and the
two public routes live in `deploy/paisa.caddy.template`.
`make deployment-test` checks the complete success and rollback orchestration.

Deploy through the infrastructure repository:

```bash
cd /path/to/hetzner_tf
./scripts/deploy-vps-prod-from-tf.sh paisa main
```

The deploy script:
- verifies that the shared VPS platform is ready
- takes a Paisa-only deployment lock
- installs the tested Compose manifests under `/opt/paisa`
- writes `/etc/paisa/app.env`
- pulls the immutable API and web images
- starts only the `paisa` Compose project
- health-checks the private services
- installs only `/opt/shared-caddy/apps/paisa.caddy` and safely reloads Caddy

GitHub Actions builds both images once, exercises them together in ephemeral
Docker staging, and deploys those exact digests. The frontend receives its API
URL at container startup rather than at image-build time.

## Runtime

Production compose services:
- `paisa-web`: React console on the `vps-edge` network as `paisa-web:80`
- `paisa-api`: Go API on the `vps-edge` network as `paisa-api:8080`
- `paisa-postgres`: private Postgres with the persistent
  `paisa_paisa-postgres-data` volume

No Paisa container publishes a host port. Shared Caddy is the only public entry
point and reaches the stable network aliases above.

The API reads `/app/db/schema.sql` and applies the schema during startup.

## Verify

```bash
curl -f https://api.paisa.example.com/health
curl -I https://paisa.example.com
sudo docker compose -p paisa --env-file /etc/paisa/app.env \
  -f /opt/paisa/compose.yml -f /opt/paisa/compose.production.yml ps
sudo docker compose -p paisa --env-file /etc/paisa/app.env \
  -f /opt/paisa/compose.yml -f /opt/paisa/compose.production.yml logs --tail=100 paisa-api
```
