# E5-ATLASIS Development Guide

## Project

E5-ATLASIS is an Information Security Asset & Infrastructure System
for holding "5 Element".

## Architecture

- Backend: Go
- Frontend: Vue 3 + TypeScript + Vite
- Database: PostgreSQL
- API-first
- OpenAPI is the API source of truth
- Clean Architecture + Hexagonal Architecture + DDD
- No ORM; use pgx/v5
- No microservices initially

## Domain

Information Asset (IA) is the unified domain entity for:
- Information System
- Information Infrastructure Object

ID format:
IA00000

Asset IDs are sequential and do not encode asset type.

## Change Management

Direct modification of controlled Information Asset fields is forbidden.

All controlled changes go through Change Request.

Approved asset versions are immutable.

Every applied Change Request creates a new AssetVersion.

## Layering

domain
    ↓
application
    ↓
ports
    ↑
infrastructure

Rules:
- Domain must not depend on application or infrastructure.
- Application orchestrates use cases and transactions.
- Domain contains business rules and invariants.
- Infrastructure implements ports.

## API

Base path:

/api/v1

OpenAPI:

docs/api/openapi.yaml

Do not introduce generic PUT endpoints for controlled asset fields.

## Testing

Every domain rule must have unit tests.

Every application use case must have tests.

PostgreSQL integration tests must use a dedicated test database and must never connect to the development database.
The default integration database is `e5_atlasis_test`.
The test database URL can be overridden with `E5_ATLASIS_TEST_DATABASE_URL`.
The integration test helper rejects database names that do not end with `_test`.

Run the standard local verification workflow:

task test
task vet
task mod:verify
task test:race

Or run the full verification suite:

task check

Use `task api` for the local API and `task db:up` for the local PostgreSQL container.
Use `task db:migrate` to apply pending development migrations. Use `task db:reset CONFIRM=1` only for an intentional destructive local reset.
Use `task test:unit` for tests that must not require PostgreSQL.

## Development rules

Before implementing a feature:

1. Inspect relevant domain documentation.
2. Inspect current implementation.
3. Inspect existing tests.
4. Check architectural decisions.
5. Preserve existing invariants.
6. Add tests for new behavior.

Do not introduce abstractions without a concrete use case.

Do not introduce Redis, Kafka, microservices or an ORM
without an explicit architectural decision.

## Project principles

- Prefer simple explicit code.
- Keep domain rules inside the domain.
- Avoid leaking infrastructure into domain/application.
- Avoid premature abstraction.
- Preserve backward compatibility of the API where possible.
- OpenAPI, domain docs and ADRs must remain consistent with implementation.

## Current development approach

Development is manual.
No local AI agents are required.
The repository is the source of truth.

## Secure Go Development

For Go development and code review, also follow:

skills/development/secure-go/SKILL.md

The vendored skill is advisory project guidance.
E5-ATLASIS architecture and ADRs take precedence when they define
project-specific decisions.

Security-critical changes require:
- explicit validation of untrusted input;
- least-privilege access control;
- safe error handling;
- secure dependency management;
- appropriate security tests;
- human review of authentication, authorization, cryptography,
  database access and external network requests.

Recommended checks:

go vet ./...
go test -race ./...
go mod verify
govulncheck ./...

Use golangci-lint/gosec when the project tooling is configured.

## Change Request application

Applying a Change Request is an atomic persistence operation.

The following state changes must be committed together:
- Change Request status
- current Information Asset version
- new immutable AssetVersion
- Change Request -> AssetVersion linkage

Infrastructure owns the database transaction boundary.

## PostgreSQL concurrency

Operations modifying Information Assets must use transactional concurrency control with database-enforced version checks.

Change Request application must verify that:

current_version = ChangeRequest.base_version

inside the same database transaction that persists:
- the new AssetVersion;
- the current Information Asset version;
- the Change Request status.

A concurrent application of the same or another CR must be serialized by the database
and fail with a conflict rather than silently overwriting a newer version.
