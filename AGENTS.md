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

Run:

go test ./...

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
