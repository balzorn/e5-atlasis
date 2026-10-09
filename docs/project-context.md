# AtlasIS Project Context and Collaboration Protocol

**Project:** E5-ATLASIS — Information Security Asset & Infrastructure System for holding "5 Element".  
**Last reviewed:** 2026-10-09.

This document records durable project agreements for maintainers and AI-assisted development. It does not replace current source code, migrations, tests, accepted ADRs or legal-source verification.

## 1. Source of truth

When sources disagree, do not silently choose one. Identify the conflict and propose a documentation/ADR update before implementation.

1. Explicitly approved decisions and user-confirmed constraints recorded in repository documentation.
2. Accepted ADRs, within their stated scope. Proposed ADRs are not accepted decisions or proof of implementation.
3. Current implementation: Go source, SQL migrations, tests, OpenAPI, frontend and Taskfile.
4. Architecture/domain documents describing target behavior.
5. Roadmap and implementation plans, which must be reconciled with code regularly.
6. Conversation or external notes are context only; durable decisions should be written to the repository.

AGENTS.md defines development rules. The vendored secure-go skill is advisory; project-specific architecture and accepted ADRs take precedence.

## 2. Baseline and workflow

The reviewed planning baseline is branch feat/authorization-model, commit 2cf7bc8153a08d29e9d3b87a443eb00384896f44 (2cf7bc8). At review time, main is behind this baseline. Check live branch and PR state before relying on it; do not assume the baseline has been merged into main.

The implementation plan is proposed in PR #3: https://github.com/balzorn/e5-atlasis/pull/3. Review the current PR status and actual diff before treating the plan as accepted.

Working rules:
- Before substantial planning or implementation, inspect repository guidance, relevant ADRs, domain docs, code, tests, migrations, OpenAPI and Taskfile. Do not rely only on remembered context.
- Begin with read-only inspection and state the actual branch/ref inspected.
- Prefer small, reviewable PRs. Do not merge PRs without explicit authorization.
- Do not change application code, database schema, migrations, API contracts or deployment configuration until the user has approved the design/scope.
- Architecture and migration proposals precede irreversible migrations. Never rewrite an applied migration; add a new one only after the target model is approved.
- Do not run destructive database tasks, reset databases, fetch/alter local state during diagnostics, or change task behavior unless explicitly requested.
- Report files changed, commit/PR, checks actually run and results. Never claim an unrun check passed.
- If sources conflict, report the discrepancy and propose a minimal correction.
- After an important decision is confirmed, record it in the relevant ADR/domain/architecture document and update the roadmap as needed.

## 3. Product and target domain

AtlasIS is a holding-wide registry for information systems and objects of information infrastructure, including responsible persons, purpose, criticality/risk, information protection, attestation and interaction with the cybersecurity center. Registry attributes and validations must be traced to applicable Republic of Belarus requirements, especially OAC orders No. 66 and No. 130. Do not infer legal requirements from memory; cite exact provisions in a traceable matrix.

Agreed target architecture for v1:
- IS and OII are separate domain entities with independent IDs, owners, attributes, permissions and versions.
- An OII may contain one or more IS; an IS may belong to multiple OIIs. Nested OIIs are unsupported in v1.
- OII composition does not transfer ownership, attributes or permissions.
- Composition changes go through a CR targeting the OII. Composition snapshot belongs to the immutable OII version; do not maintain a second editable current-membership source of truth.
- A CR targets exactly one entity. Multi-entity atomic CRs are out of scope for v1.
- Tenant is the isolation boundary. Reads, writes, search, counts, exports and lookups must enforce tenant isolation. Every organization belongs to exactly one tenant; typed organization relationships do not grant access by themselves.
- Role assignments have explicit scope. Tenant-wide grants are separate. Managing OII composition is a distinct permission.
- New technical entity keys use UUIDv7. Human-readable numbers are separate sequences scoped by tenant and entity type. Gaps are acceptable; identifiers are never reused and must not be calculated using COUNT(*) + 1.
- Existing IA identifiers remain migration references until a mapping strategy is approved.
- Keycloak/OIDC is the intended v1 identity boundary. Verified (iss, sub) maps to an immutable internal subject UUID. Mutable login/profile claims do not grant permissions; AtlasIS owns role assignments and their audit history.
- PostgreSQL RLS may provide defense in depth, not replace application authorization.
- Preserve immutable versions, atomic CR application, optimistic version checks and database constraints.
- Individually managed non-IS technical assets, nested OIIs and multi-resource CRs are deferred from v1.

These are target decisions, not proof of implementation. See docs/architecture/README.md, data-model-v1.md, access-control-model-v1.md and the ADRs.

## 4. Current implementation versus target

Current code still uses the legacy unified InformationAsset model and IA-style IDs. Existing migrations do not yet implement the complete tenant/organization model, distinct IS/OII persistence, OII composition persistence or a fully persisted discussion system. OIDC is not implemented.

Development authentication mode is deliberately limited: development-header must remain loopback-only and is not production authentication. The default authentication configuration is fail-closed. Do not treat X-Actor-ID as verified identity or expose development mode to other hosts. A declared RBAC vocabulary/authorizer does not prove all application use cases enforce authorization; inspect actual call sites.

Older domain and roadmap docs may describe the legacy model. When implementing target decisions, reconcile legacy docs rather than treating their IA wording or checkboxes as overriding the target architecture.

## 5. Engineering constraints

- Backend Go; frontend Vue 3 + TypeScript + Vite; PostgreSQL persistence.
- Modular monolith initially; Clean/Hexagonal Architecture and DDD boundaries.
- OpenAPI-first; docs/api/openapi.yaml is the API contract source of truth.
- Use pgx/v5; no ORM or microservices without an explicit decision.
- Preserve domain/application/infrastructure dependency direction.
- Controlled business-field changes go through CR; effective versions are immutable.
- CR application is one atomic persistence operation, including CR state, current-version pointer, new immutable version and CR-to-version link.
- Use database concurrency controls and expected-version/status checks. Preserve and extend concurrency/persistence tests.
- Prefer real foreign keys, checks, unique and not-null constraints. Revisit cascades that could destroy audit/history.
- Do not couple the domain to HTTP, OIDC or a policy engine.

## 6. Local workflow and verification

Use the existing Taskfile.yml as the standard task interface. Inspect current task definitions before recommending alternatives; preserve task behavior. Diagnostics should be read-only.

Repository verification commands include task test:unit, task test, task test:race, task vet, task mod:verify and task check.

PostgreSQL integration tests must use e5_atlasis_test or an explicitly configured database name ending in _test; never run them against the development database. task db:reset CONFIRM=1 is destructive and requires explicit intent.

The developer workflow may use OpenCode, a Windows-hosted Ollama model, RTK and project-specific guidance. The repository tree reviewed on 2026-10-09 contains AGENTS.md and the vendored secure-go skill, but no committed OpenCode agent/config files were found. If those instructions are required for reproducible work, commit the actual project-scoped configuration in a future tooling/documentation PR; do not assume a local-only setup is available to other contributors.

## 7. Security and verification

For Go work, follow skills/development/secure-go/SKILL.md and its references. Security-sensitive changes require appropriate tests and human review, especially authentication, authorization, cryptography, database access and external network requests. Recommended checks, when available, include go vet ./..., go test -race ./..., go mod verify and govulncheck ./....

Do not store secrets or credentials in the repository. Validate untrusted input, fail closed, use least privilege, avoid leaking sensitive details through errors/logs, and test both allowed and denied cases.

## 8. Planning protocol

For every substantial stage:
1. Re-read this file, AGENTS.md, the implementation plan, related ADRs and relevant code/migrations/tests.
2. State the baseline branch and commit used.
3. Separate implemented behavior from proposed target behavior.
4. List unresolved decisions and source conflicts.
5. Propose the smallest reviewable deliverable and acceptance criteria.
6. Obtain approval before code/schema/migration changes.
7. Update repository knowledge when a decision is confirmed.

The next design stage is the target data model and normative attribute/validation matrix for OAC orders No. 66 and No. 130. This stage is documentation/design only until implementation is explicitly approved.
