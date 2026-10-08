# ADR-008: Transactional concurrency control for Change Request application

## Status

Accepted

## Decision

Change Request application uses transactional concurrency control enforced by PostgreSQL.

A Change Request contains the Information Asset version from which it was created.

When applying a Change Request, the infrastructure transaction must:

1. Lock the current Information Asset row using `SELECT ... FOR UPDATE`.
2. Verify that `current_version` equals `ChangeRequest.base_version`.
3. Create the next immutable `AssetVersion`.
4. Update the current Information Asset version using the expected base version.
5. Transition the Change Request from `APPROVED` to `APPLIED`.
6. Commit all state changes in a single transaction.

The version check and all state changes above must occur within the same database transaction.

A concurrent application of the same or another Change Request against the same base version is serialized by the database row lock. After the first transaction commits, a competing transaction observes the new `current_version` and fails the base-version check.

The competing operation must roll back without creating an additional AssetVersion or changing the Change Request state.

No external or distributed locking mechanism is required at this stage.

## Rationale

This prevents:

- lost updates;
- applying a stale Change Request;
- two Change Requests producing the same next asset version;
- partial persistence of an applied Change Request.

The database remains the authority for concurrency control at the persistence boundary.
