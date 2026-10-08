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

- Information Assets
- Change Requests
- Discussions
- Approvals

## Key principle

Controlled Information Asset changes are performed through
Change Requests and produce immutable Asset Versions.