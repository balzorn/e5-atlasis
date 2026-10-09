# E5-ATLASIS Development Guide

## Project

E5-ATLASIS is an Information Security Asset & Infrastructure System for holding "5 Element".

Read docs/project-context.md for durable project context, collaboration rules, current-versus-target distinctions and the planning protocol.

## Technology and architecture

- Backend: Go.
- Frontend: Vue 3 + TypeScript + Vite.
- Database: PostgreSQL.
- API-first; docs/api/openapi.yaml is the API contract source of truth.
- Clean Architecture + Hexagonal Architecture + DDD, implemented as a modular monolith.
- Use pgx/v5; no ORM or microservices without an explicit architecture decision.

Layering:

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
- Do not couple the domain to HTTP, OIDC or a policy engine.

## Domain: current implementation versus target

The current code has a legacy unified InformationAsset entity with IAxxxxx identifiers and a type attribute. This describes the existing implementation, not the agreed target model.

The target v1 architecture treats Information Systems (IS) and Objects of Information Infrastructure (OII) as distinct domain entities with independent IDs, owners, attributes, permissions and versions. An IS may belong to multiple OIIs; nested OIIs are out of scope. OII composition does not transfer ownership or permissions and changes through a CR targeting the OII. See docs/architecture/README.md, docs/architecture/data-model-v1.md and docs/project-context.md.

For target identifiers, technical keys use UUIDv7 and human-readable numbers are separate sequences scoped by tenant and entity type. Existing IA identifiers are retained as migration references until the mapping strategy is approved. Do not implement this target or create migrations until the data model and migration plan are approved.

## Change management and persistence invariants

- Direct modification of controlled fields is forbidden; use Change Requests.
- Effective versions are immutable.
- Every applied CR creates a new version.
- Applying a CR is one atomic persistence operation: CR status, current-version pointer, new immutable version and CR-to-version link commit together.
- Use database-enforced concurrency checks, expected base-version checks and appropriate row locks.
- Preserve transaction, versioning, status-transition and concurrency guarantees and tests.
- Prefer foreign keys, CHECK, UNIQUE and NOT NULL constraints. Review delete cascades where they may destroy audit/history.
- Never use COUNT(*) + 1 for identifier allocation.

## Tenant and authorization target

- Tenant is the isolation boundary; enforce it for reads, writes, search, counts, exports and lookups.
- Each organization belongs to exactly one tenant. Organization relationships do not automatically grant permissions.
- Role assignments have explicit scope; tenant-wide grants are separate.
- Permissions do not inherit through OII composition; managing composition is a separate action.
- Keycloak/OIDC is the intended v1 identity boundary. Map verified (iss, sub) to an immutable internal subject UUID; mutable login, email, UPN, domain and display-name attributes are not authorization grants.
- PostgreSQL RLS, if adopted, is defense in depth rather than a replacement for application authorization.

These are target principles, not a claim that tenant isolation, OIDC or the full authorization model is implemented.

## API

Base path: /api/v1.
OpenAPI: docs/api/openapi.yaml.

Do not introduce generic PUT endpoints for controlled business fields. Do not trust actor identity supplied in request parameters or headers in production.

The development-header identity mode is not authentication and must remain loopback-only. The default authentication configuration is fail-closed. OIDC integration is not yet implemented.

## Testing and local tasks

Use the existing Taskfile.yml as the standard local task runner. Inspect task definitions before suggesting or changing workflows. Diagnostics must be read-only and must not fetch or alter repository/database state. Preserve existing task behavior.

Typical verification commands:

    task test:unit
    task test
    task vet
    task mod:verify
    task test:race
    task check

- Every domain rule must have unit tests.
- Every application use case must have tests.
- Cover success, invalid input, forbidden actions, conflicts and relevant concurrency cases.
- PostgreSQL integration tests must use the dedicated e5_atlasis_test database (or a configured name ending in _test) and never the development database.
- task db:reset CONFIRM=1 is destructive and must only be run with explicit intent.
- Never store real credentials in the repository.

## Development rules

Before implementing a feature:

1. Read docs/project-context.md and relevant domain documentation.
2. Inspect the actual branch/ref, implementation, migrations and tests.
3. Read related ADRs and distinguish Accepted from Proposed.
4. Check the OpenAPI contract and Taskfile where relevant.
5. Preserve invariants and backward compatibility where possible.
6. Propose design and acceptance criteria before substantial code/schema changes.
7. Add tests and update docs/roadmap.md when behavior or scope changes.
8. Report exactly which checks were run and their results.

Do not introduce abstractions without a concrete use case. Do not introduce Redis, Kafka, microservices, an ORM or a new policy engine without an explicit architectural decision.

## Secure Go Development

For Go development and code review, follow skills/development/secure-go/SKILL.md. It is advisory guidance; E5-ATLASIS architecture and accepted ADRs take precedence when they define project-specific decisions.

Security-critical changes require input validation, least-privilege access control, safe error handling, secure dependency management, appropriate security tests and human review of authentication, authorization, cryptography, database access and external network requests.

Recommended checks when available:

    go vet ./...
    go test -race ./...
    go mod verify
    govulncheck ./...

Use configured golangci-lint/gosec checks when available.

## Change and planning workflow

- Repository documentation is the durable project source of truth; record confirmed decisions in relevant docs/ADRs.
- Before substantial planning, inspect repository knowledge and identify conflicts or stale docs rather than relying on conversational memory.
- Prefer small PRs and explicit acceptance criteria.
- Do not merge PRs or implement unapproved architectural changes.
- Do not modify application code, database schema, migrations, API contracts or deployment configuration before explicit approval of the proposed design/scope.
- Never rewrite an applied migration; add a new migration only after the target schema is approved.
- If docs conflict with code or each other, report the discrepancy and propose the smallest corrective update.
