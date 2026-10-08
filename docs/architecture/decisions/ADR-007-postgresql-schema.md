# ADR-007: PostgreSQL persistence model

## Status

Accepted

## Decision

PostgreSQL stores the current Information Asset pointer separately from
immutable Information Asset versions.

### Information Asset

`information_assets` contains:

- IA identity
- current version pointer

### Information Asset Version

`information_asset_versions` contains a complete immutable snapshot.

### Change Request

`change_requests` stores workflow state and references the asset and base version.

### Field Changes

`change_request_changes` stores proposed values as typed JSONB objects.

The application/domain layer remains responsible for semantic validation.

### Approvals

`approvals` stores explicit approval decisions associated with a Change Request.

## Invariants enforced by PostgreSQL

- IA identifier format
- version >= 1
- valid enumerated values
- Change Request base version must exist
- current IA version must exist
- one AssetVersion may be produced by at most one Change Request
- field changes contain different old/new values

## Transaction

Application of a Change Request is persisted in one database transaction.
