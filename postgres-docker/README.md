```
./setup.sh
docker exec -it $(docker-compose ps -q postgres) psql -U project -d project
docker-compose down -v
```