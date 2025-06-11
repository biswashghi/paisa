docker-compose up -d

# Wait for Postgres to be ready
until docker exec $(docker-compose ps -q postgres) pg_isready -U project -d project; do
  echo "Waiting for postgres to be ready..."
  sleep 2
done

echo "Postgres is ready and tables should be initialized."