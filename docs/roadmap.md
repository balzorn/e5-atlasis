# E5-ATLASIS Roadmap

This roadmap distinguishes the current legacy implementation from the agreed target architecture. Checkboxes describe repository state, not merely the existence of domain types or design documents. See docs/project-context.md and docs/architecture/implementation-plan-v1.md.

## Current implementation baseline

### Domain and workflow
- [x] Legacy unified Information Asset model (IS/OII represented by a type attribute)
- [x] Reference values and asset creation
- [x] Change Request domain and creation
- [x] Change Request submission and review workflow
- [x] Approval model and approval decisions
- [x] Change Request application use case
- [x] Immutable AssetVersion persistence
- [x] Atomic CR application and database concurrency controls
- [x] Database-enforced CR transition concurrency
- [x] Domain discussion types
- [ ] Persistent discussion threads/comments and edit-history persistence

### Backend and API
- [x] PostgreSQL schema and migrations for the legacy model
- [x] pgx repositories for the current schema
- [x] HTTP REST API for current endpoints
- [x] OpenAPI contract and contract tests
- [x] Fail-closed default authentication behavior
- [x] Loopback-only development-header identity mode
- [ ] Production authentication / Keycloak OIDC
- [ ] Full tenant-aware authorization enforced by all application use cases
- [ ] Tenant/organization catalogues and typed organization relationships
- [ ] Separate IS and OII persistence models
- [ ] Tenant- and type-scoped human-readable IS/OII identifiers
- [ ] OII-to-IS composition persistence and version snapshots
- [ ] Persistent audit/event history
- [ ] Normative attribute validation matrix for OAC orders No. 66 and No. 130
- [ ] Compliance policy execution (policy engine and integration approach to be decided)
- [ ] OpenTelemetry integration

### Frontend
- [ ] Functional IS registry and list/search
- [ ] IS details and version history
- [ ] OII details and composition management
- [ ] CR create/edit/submit workflow
- [ ] Approval and review workflow UI
- [ ] Persistent discussions UI
- [ ] Administration and scoped role assignment

### Infrastructure
- [ ] Production deployment model and operational documentation
- [ ] Attachments storage (S3/MinIO) if confirmed as a v1 requirement
- [ ] Directory/identity integration beyond the Keycloak OIDC boundary, if required

## Target v1 sequence

1. [ ] Approve the implementation plan and repository operating guidance.
2. [ ] Design the target data model and normative attribute/validation matrix.
3. [ ] Approve legacy IA-to-IS/OII mapping and data migration strategy.
4. [ ] Implement tenant and organization foundations plus distinct IS/OII models.
5. [ ] Adapt CR, immutable versions, approvals, OII composition and discussions/audit.
6. [ ] Implement verified OIDC identity mapping and tenant-aware authorization.
7. [ ] Finalize API compatibility and implement functional registry UI.
8. [ ] Run migration rehearsal, security/concurrency regression tests and cutover rehearsal.

## Explicitly deferred unless separately approved

- Redis
- Kafka
- Microservices
- ORM
- Individually managed non-IS technical assets
- Nested OIIs
- Multi-entity atomic CRs
- Local AI agent infrastructure as a production runtime dependency
