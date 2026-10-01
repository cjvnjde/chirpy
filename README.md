Connect to the database using psql

```bash
psql postgres://postgres:postgres@localhost:5432/chirpy
```

Start postgres in docker

```bash
docker run -d \
  --name pg \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  -v pgdata:/var/lib/postgresql \
  postgres:18
```

```bash
docker start pg
```
```bash
docker stop pg
```
Create database
```bash
docker exec pg psql -U postgres -d postgres -c 'CREATE DATABASE chirpy;'
```
Open sql inside docker
```bash
docker exec -it pg psql -U postgres -d chirpy
```
or usinng psql
```bash
psql "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
```

run migration up
```bash
goose -dir sql/schema postgres "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable" up
```
or down
```bash
goose -dir sql/schema postgres "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable" down
```
