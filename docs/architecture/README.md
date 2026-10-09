# Architecture

E5-ATLASIS is initially designed as a modular monolith with explicit domain boundaries.

## Main layers

### Domain
Business entities, value objects and invariants.

### Application
Use cases, orchestration and application-level policies.

### Ports
Interfaces required by the application.

### Infrastructure
PostgreSQL, HTTP, authentication, external systems and other adapters.

## Main domains

- Tenants, organizations and departments
- Information Systems (IS)
- Objects of Informatization (OII)
- OII composition and composition history
- Change Requests, approvals and discussions
- Identity mappings, role assignments and audit history

Individually managed non-IS technical assets are deferred from v1. OII composition in v1 links OIIs to ISs only.

## Identity and access

Keycloak is the v1 OIDC identity provider. AtlasIS maps a stable external identity to an internal UUID subject, maintains role assignments and authorization scope itself, and must not treat mutable login, email or organization claims as the internal primary key or as grants of permissions.

Corporate Active Directory is the initial authoritative directory for user identity attributes. The v1 integration boundary receives verified claims from Keycloak; direct AD synchronization and a future corporate identity API are integration options, not prerequisites for the internal authorization model.

## Key principles

- IS and OII are distinct entities with independent identities, owners and versions.
- An IS may belong to multiple OIIs; nested OIIs are not supported in v1.
- Ownership, attributes and access do not inherit across composition links.
- Human-readable numbers use a prefix and a sequence scoped by tenant and entity type. Database relations use UUIDv7.
- Each Change Request targets exactly one IS or OII. An OII composition change targets the OII.
- Controlled changes produce immutable effective versions and retain an auditable history.

See [Data Model v1](data-model-v1.md), [Access Control Model v1](access-control-model-v1.md) and [ADR-011](decisions/ADR-011-authorization-model.md) for the detailed proposal.

These documents specify the proposed architecture. They do not mean the schema, OIDC integration or authorization model has already been implemented.
