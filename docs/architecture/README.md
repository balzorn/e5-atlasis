# Architecture

E5-ATLASIS is initially a modular monolith.

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

- Information Systems
- Objects of Informatization
- Technical Assets (where individually managed)
- Change Requests
- Discussions
- Approvals

See [Data Model v1](data-model-v1.md) for the proposed relational model and unresolved design decisions.

## Key principle

Controlled Information Asset changes are performed through
Change Requests and produce immutable Asset Versions.