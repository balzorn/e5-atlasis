# ADR-010: Change Request status transitions use database compare-and-set

## Status

Accepted

## Decision

Change Request status updates are persisted with an expected previous status.

The repository update must match both:
- the Change Request identifier;
- the expected current status.

Conceptually:

`UPDATE change_requests SET status = new_status WHERE id = ? AND status = expected_status`

If exactly one row is not affected, the operation fails with a conflict.

Application use cases capture the status before the domain transition and pass it to the persistence port.

## Rationale

A Change Request can be acted on concurrently by multiple clients. Reading the same status and then performing an unconditional update would allow a stale operation to overwrite a newer transition.

Database-enforced compare-and-set provides:
- deterministic concurrency control;
- no distributed lock;
- correct HTTP 409 conflict semantics;
- consistency with approval and asset concurrency controls.

## Consequences

Concurrent transitions of the same Change Request are serialized by PostgreSQL.

Exactly one operation can successfully consume a given expected status.

The application layer remains responsible for defining valid domain transitions; PostgreSQL enforces that the persisted transition is still based on the state that the application observed.
