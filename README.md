# E5-ATLASIS

E5-ATLASIS is an Information Security Asset & Infrastructure System for holding "5 Element".

## Development workflow

The project uses [Task](https://taskfile.dev/) as the standard local task runner.

Install Task using one of the official methods documented at https://taskfile.dev/docs/installation.

From the repository root:

```bash
task --list
```

Common commands:

```bash
task db:up
task db:migrate
task fmt
task test:unit
task test
task test:postgres
task test:race
task vet
task mod:verify
task check
task api
```

The default development PostgreSQL database is `e5_atlasis`.

Use `task db:migrate` to apply pending migrations to the development database. Migration history is stored in `schema_migrations`.

For a clean local development database, stop the API first and run:

```bash
task db:reset CONFIRM=1
```

`db:reset` is intentionally destructive and recreates the local `public` schema before applying all migrations.

PostgreSQL integration tests use a separate `e5_atlasis_test` database and must never run against the development database. The integration test harness creates the test database and applies the repository migrations automatically.

To override the integration database:

```bash
E5_ATLASIS_TEST_DATABASE_URL='postgres://user:password@host:5432/my_project_test?sslmode=disable' task test
```

Do not store real credentials in the repository. The URLs in the local Taskfile are development-only defaults matching the bundled local Compose environment.


## API identity during development

The `task api` task enables the explicit `development-header` identity mode for local development.
In this mode, requests must include `X-Actor-ID`, and the API process refuses to start unless
`HTTP_ADDR` is a loopback IP address such as `127.0.0.1:8080` or `[::1]:8080`.

This header is **not authentication** and grants no roles or organization permissions. Do not expose
this mode to other hosts or use it in production. When `ATLASIS_AUTH_MODE` is unset, the API starts
in fail-closed mode and rejects all `/api/v1/` requests with HTTP 401 until a trusted authentication
resolver is configured. OIDC integration is not implemented yet.
