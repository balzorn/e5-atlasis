# Change Request

A Change Request (CR) is the only mechanism for changing controlled fields of an Information Asset.

## Identity

- CR ID is unique.
- CR belongs to exactly one Information Asset.
- CR contains the base asset version against which the change was created.

## Statuses

- DRAFT
- SUBMITTED
- UNDER_REVIEW
- CHANGES_REQUESTED
- APPROVED
- APPLYING
- APPLIED
- REJECTED
- CANCELLED

## Change

Each changed field is represented separately:

- change ID
- field
- old value
- new value

The server derives old values from the asset version.

## Invariants

1. CR cannot be applied against a different asset version.
2. APPROVED CR can be applied only once.
3. APPLYING and APPLIED are system-controlled states.
4. Every applied CR creates a new immutable AssetVersion.
5. Direct modification of controlled asset fields is forbidden.
6. A Change Request cannot become APPROVED while any required Approval is not APPROVED.
7. A Change Request must have at least one Approval before it can become APPROVED.

## Lifecycle

DRAFT -> SUBMITTED

SUBMITTED -> UNDER_REVIEW
SUBMITTED -> CANCELLED

UNDER_REVIEW -> CHANGES_REQUESTED
UNDER_REVIEW -> APPROVED
UNDER_REVIEW -> REJECTED

CHANGES_REQUESTED -> UNDER_REVIEW
CHANGES_REQUESTED -> CANCELLED

APPROVED -> APPLYING
APPLYING -> APPLIED

REJECTED -> no transitions
CANCELLED -> no transitions
APPLIED -> no transitions
