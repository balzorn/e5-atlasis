# Reference Values

## Criticality

- LOW
- MEDIUM
- HIGH
- CRITICAL

## Risk Level

- LOW
- MEDIUM
- HIGH
- CRITICAL

## Protection Status

- NOT_REQUIRED
- REQUIRED
- IN_PROGRESS
- IMPLEMENTED

## Attestation Status

- NOT_REQUIRED
- REQUIRED
- IN_PROGRESS
- ATTESTED
- EXPIRED

## Lifecycle Transitions

DRAFT -> ACTIVE
DRAFT -> RETIRED

ACTIVE -> SUSPENDED
ACTIVE -> RETIRED

SUSPENDED -> ACTIVE
SUSPENDED -> RETIRED

RETIRED -> no transitions

## Rules

- DRAFT is not considered operational.
- ACTIVE is the normal operational state.
- SUSPENDED means the asset exists but is temporarily not operational.
- RETIRED is terminal.
- Lifecycle transition rules are enforced by the domain model.
