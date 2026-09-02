# VPS deployment

Paisa is one independent release unit on the shared VPS platform. Its Postgres
container stays on the private Paisa network. Only `paisa-web` and `paisa-api`
join the platform-owned `vps-edge` network.

The platform must already be bootstrapped from the infrastructure repository:

```bash
./scripts/deploy-vps-platform.sh deploy <server-ip> admin@example.com
```

Deploy Paisa from this repository or its `main` workflow:

```bash
PAISA_WEB_DOMAIN=paisa.example.com \
PAISA_API_DOMAIN=api.paisa.example.com \
PAISA_POSTGRES_PASSWORD=... \
PAISA_INTERNAL_ADMIN_TOKEN=... \
PAISA_API_IMAGE=ghcr.io/owner/repo/paisa-api@sha256:... \
PAISA_WEB_IMAGE=ghcr.io/owner/repo/paisa-web@sha256:... \
./scripts/deploy-vps.sh deploy <server-ip>
```

The deploy uses the named `paisa` Compose project, preserves the existing
`paisa_paisa-postgres-data` volume, waits for the API, then atomically updates
only `/opt/shared-caddy/apps/paisa.caddy` through `vps-route`.

CI accepts provider-neutral `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, and pinned
`VPS_KNOWN_HOSTS` secrets.
The old `HETZNER_*` names remain temporary fallbacks during migration.

Use `make local-up`, `make local-test`, and `make local-down` for the complete
local stack. `make staging-test` runs the disposable production-image gate used
by GitHub Actions.
