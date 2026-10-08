# ADR-006: Atomic Change Request application

## Status

Accepted

## Context

Applying an approved Change Request modifies several related pieces of state:

1. Change Request status
2. Current Information Asset version
3. New immutable AssetVersion
4. Change Request linkage to the new AssetVersion

A partial update would leave inconsistent state.

## Decision

Application of a Change Request is a single atomic persistence operation.

The operation must either:

- persist all related changes; or
- persist none of them.

The `APPLYING` status is an internal transient state.

A successful transaction results in:

```text
Change Request -> APPLIED
Information Asset -> current_version = N+1
AssetVersion -> N+1 created
```
A failed transaction leaves the previous consistent state unchanged.

## Transaction boundary

The transaction boundary belongs to the infrastructure adapter implementing
the `ports.ChangeApplier` port.

The application layer prepares and validates the resulting state.
The infrastructure layer provides atomic persistence.
