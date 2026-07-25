if [ -z "$PAISA_POSTGRES_PASSWORD" ]; then
  echo "Set PAISA_POSTGRES_PASSWORD before starting local Postgres."
  exit 1
fi

docker-compose up -d

# Wait for Postgres to be ready
until docker exec $(docker-compose ps -q postgres) pg_isready -U "${PAISA_POSTGRES_USER:-project}" -d "${PAISA_POSTGRES_DB:-project}"; do
  echo "Waiting for postgres to be ready..."
  sleep 2
done

echo "Postgres is ready and tables should be initialized."
