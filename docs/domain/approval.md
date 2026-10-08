# Approval

An Approval is an explicit decision required for a Change Request.

## Types

- ASSET_OWNER
- SECURITY
- ORGANIZATION
- SYSTEM_OWNER

## Statuses

- PENDING
- APPROVED
- REJECTED

## Rules

1. Each approval belongs to exactly one Change Request.
2. A required approval must be APPROVED before the Change Request can become APPROVED.
3. Only the assigned approver can make the decision.
4. A decision can only be made while the approval is PENDING.
5. An approval decision is immutable.
6. The decision timestamp is recorded.
7. A rejection does not automatically reject the Change Request.
8. Change Request rejection is an explicit workflow action.
