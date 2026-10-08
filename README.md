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
task fmt
task test
task test:postgres
task test:race
task vet
task mod:verify
task check
task api
```

The default development PostgreSQL database is `e5_atlasis`.

PostgreSQL integration tests use a separate `e5_atlasis_test` database and must never run against the development database. The integration test harness creates the test database and applies the repository migrations automatically.

To override the integration database:

```bash
E5_ATLASIS_TEST_DATABASE_URL='postgres://user:password@host:5432/my_project_test?sslmode=disable' task test
```

Do not store real credentials in the repository. The URLs in the local Taskfile are development-only defaults matching the bundled local Compose environment.
