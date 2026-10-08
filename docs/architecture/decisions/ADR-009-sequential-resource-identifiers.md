# ADR-009: Sequential resource identifiers use PostgreSQL sequences

## Status

Accepted

## Decision

Information Asset identifiers (`IAxxxxx`), Change Request identifiers (`CRxxxxx`), and Approval identifiers (`APRxxxxx`) are allocated by PostgreSQL sequences.

The application layer depends on identifier-generator ports and does not calculate the next identifier itself.

The supported identifier ranges are:

- Information Assets: `IA00001` through `IA99999`
- Change Requests: `CR00001` through `CR99999`
- Approvals: `APR00001` through `APR99999`

Identifiers are unique and monotonically increasing within the sequence. They are **not guaranteed to be gapless**.

A sequence value may be consumed when a transaction or request later fails. Such gaps are acceptable because identifiers are technical stable references, not business counters.

At application startup, no in-memory identifier state is maintained.

Database migrations initialize sequences from the highest existing identifier where compatible records already exist.

## Rationale

PostgreSQL sequences provide:

- safe concurrent allocation;
- simple persistence-boundary ownership;
- no application-side locking;
- predictable formatting;
- independence from HTTP/application instance count.

Gapless allocation would require additional transactional serialization and would add complexity without providing business value for technical identifiers.

## Consequences

Clients must never supply `IAxxxxx`, `CRxxxxx`, or `APRxxxxx` identifiers for creation.

The API returns the generated identifier after successful creation.

The current five-digit identifier format limits each identifier namespace to 99,999 values. Increasing that limit later requires a deliberate domain/API migration.