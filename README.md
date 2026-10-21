# High Load Sandbox — Social Network Skeleton

Monolithic social-network backend written in plain Go (standard library
`net/http` only, no web framework) with PostgreSQL accessed through
parameterized raw SQL (no ORM, no query builder).

## Features

- `POST /login` — authenticate with login/password, returns a signed JWT.
- `POST /user/register` — create a user profile (first name, last name, birth
  date, gender, interests, city) with a bcrypt-hashed password.
- `GET /user/get/{id}` — fetch a user profile by ID. Requires a valid JWT from
  `/login`, sent as `Authorization: Bearer <token>`.

Passwords are hashed with bcrypt before storage. All SQL statements use
placeholder parameters ($1, $2, ...), so the app is not vulnerable to SQL
injection.

## Project layout

```
cmd/server        entry point (main.go)
internal/config    environment-based configuration
internal/db        database/sql + pgx driver connection setup
internal/model     domain types
internal/repository raw parameterized SQL queries (no ORM)
internal/handler   HTTP handlers
internal/auth      password hashing (bcrypt) and JWT issuing/validation
migrations         SQL schema, auto-applied by the postgres container on first start
bruno              Bruno collection for manual testing
```

## Running locally with Docker

Requirements: Docker and Docker Compose, Go 1.27+ (to vendor dependencies before the build —
the corporate network's TLS-inspecting proxy breaks `go mod download` from inside containers).

```bash
go mod vendor  # regenerates the vendor/ directory used by the Docker build; not committed to git
docker compose up --build
```

This starts:
- `postgres` on `localhost:5432` (db `high_load_sandbox`, user/password `postgres`/`postgres`),
  automatically applying `migrations/0001_init.sql` on first start.
- `app` (the Go server) on `localhost:9080`.

## Running locally without Docker

Requirements: Go 1.27+, a running PostgreSQL instance.

```bash
# apply the schema once
psql "$DATABASE_URL" -f migrations/0001_init.sql

# run the server
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/high_load_sandbox?sslmode=disable"
export HTTP_ADDR=":9080"
export JWT_SECRET="dev-secret-change-me"
go run ./cmd/server
```

## Example requests

```bash
curl -X POST localhost:9080/user/register \
  -H 'Content-Type: application/json' \
  -d '{"login":"ivan","password":"secret123","first_name":"Ivan","last_name":"Ivanov","birth_date":"1990-01-01","gender":"male","interests":"go, chess","city":"Moscow"}'

curl -X POST localhost:9080/login \
  -H 'Content-Type: application/json' \
  -d '{"login":"ivan","password":"secret123"}'

curl localhost:9080/user/get/1 \
  -H 'Authorization: Bearer <token-from-login-response>'
```

## Bruno collection

Open `bruno/` as a collection in [Bruno](https://www.usebruno.com/) (Open Collection →
select the `bruno` folder). It contains requests for all three endpoints,
with the registration request's response ID and the login token chained into
the "Get user" and subsequent requests via runtime variables (`userId`,
`token`). Select the "Local" environment to set `baseUrl`.

## Non-functional notes

This is a deliberately bare-bones skeleton for a high-load engineering
course: no ORM, no caching, no replication, no sharding, and no indexes
beyond what the primary/unique keys require. These concerns are introduced
and optimized in subsequent homework assignments.
