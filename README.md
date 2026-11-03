# Nebula Platform Backend

Go/Gin backend with OAuth login, PostgreSQL sessions, and a React test frontend.

## Prerequisites

Install these tools before starting:

- Go 1.26.3 or newer (see `go.mod`).
- Node.js 22 or newer, including npm.
- Docker with Docker Compose v2, supporting `up --wait`. Start Docker Desktop
  or your Docker daemon before setup.
- [just](https://github.com/casey/just), which runs the repository's development commands.
- [Goose](https://pressly.github.io/goose/), which applies the SQL migrations.

On macOS, install just with Homebrew:

```sh
brew install just
```

Install the pinned Goose CLI and make sure Go's binary directory is on your PATH:

```sh
go install github.com/pressly/goose/v3/cmd/goose@v3.28.0
export PATH="$(go env GOPATH)/bin:$PATH"
```

Run the commands below from the repository root. Docker runs PostgreSQL; the
backend and test frontend run on your machine.

## 1. Configure the environment

Copy `.env.template` to `.env` if `.env` does not already exist. Keep an existing
`.env` so you do not overwrite local credentials. Alternatively, `just setup`
creates the file if it is missing and stops so you can fill it in.

Get the Google OAuth client ID and secret from the project maintainer through the
team's private channel, then put them in the root `.env`:

```dotenv
PORT=8081
POSTGRES_PORT=5433
DATABASE_URL=postgres://nebula:nebula_dev@localhost:5433/nebula_platform?sslmode=disable

GOOGLE_CLIENT_ID=<provided by the maintainer>
GOOGLE_CLIENT_SECRET=<provided by the maintainer>
GOOGLE_REDIRECT_URL=http://localhost:8081/api/auth/google/callback
```

The maintainer must register that exact redirect URL on the Google OAuth client.
If the Google app is restricted to test users, the maintainer must also add your
Google account before you can sign in.

`.env` is ignored by Git. Keep the secret on the backend; the frontend does not
need it or its own environment file. Existing shell environment variables take
precedence over `.env`, so unset old values if the app uses unexpected settings.

## 2. Install dependencies and prepare the database

```sh
just setup
```

This downloads Go dependencies, installs the locked frontend dependencies with
`npm ci`, starts the `postgres` service from `docker-compose.yml`, waits up to 60
seconds for its health check, and runs all pending Goose migrations.

Setup can be rerun: it preserves `.env` and the Postgres volume, while Goose
applies only pending migrations. It runs migrations against `DATABASE_URL`, so
use the local development URL above.

If you prefer to run the steps individually:

```sh
go mod download
npm --prefix test-frontend ci
just db-up
just migrate-up
```

## 3. Start the backend and frontend

In one terminal:

```sh
just run
```

In a second terminal:

```sh
just frontend
```

Open [the test frontend](http://localhost:3000) and use Google login. The frontend
proxies `/api` requests to the backend, and OAuth redirects use the backend directly.

| Service | Local URL |
| --- | --- |
| Test frontend | http://localhost:3000 |
| Backend health | http://localhost:8081/health |
| Swagger UI | http://localhost:8081/swagger/index.html |

The test frontend and auth redirects currently expect frontend port 3000 and
backend port 8081. The frontend command fails if port 3000 is occupied rather than
silently choosing another port. Use `localhost` consistently, including for login.

New accounts have the `user` role. The admin users page requires an account with
the `admin` role; normal users receive HTTP 401 from the admin endpoint.

## Database and documentation commands

```sh
just db-status       # Show the database container status
just db-logs         # Follow database logs
just migrate-status  # Show applied and pending migrations
just migrate-up      # Apply pending migrations
just db-shell        # Open psql in the database container
just docs            # Regenerate Swagger documentation
just test            # Run Go tests
```

For database connection failures, check that Docker is running and `just db-status`
shows a healthy Postgres service. If port 5433 is occupied, change both
`POSTGRES_PORT` and the port in `DATABASE_URL` to the same free port, then rerun
setup. Database defaults in Compose must match the username, password, and database
in your local `DATABASE_URL`.

See [API documentation](docs/README.md) for Swagger details.

## Stop local development

Press Ctrl+C in each application terminal, then stop Postgres:

```sh
just db-down
```

This keeps database data in its Docker volume. Do not add `--volumes` unless you
intend to delete the local database.
