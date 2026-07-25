```
export PAISA_POSTGRES_PASSWORD="<local-dev-password>"
./setup.sh
docker exec -it $(docker-compose ps -q postgres) psql -U "${PAISA_POSTGRES_USER:-project}" -d "${PAISA_POSTGRES_DB:-project}"
docker-compose down -v
```

Postgres initializes from `../db/schema.sql`, which is also used by the Go API during local startup.
