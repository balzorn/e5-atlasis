# ADR-011: Authorization Model

- Status: Proposed
- Date: 2026-10-08

## Context

The current API reads the caller identifier from the `X-Actor-ID` request header.
This is a temporary development placeholder, not authentication: an arbitrary client can spoof it
unless a trusted, authenticated gateway removes any client-supplied `X-Actor-ID` header and injects
a verified actor identifier. The API must not rely on this header as a production security boundary
or expose the current configuration directly to untrusted clients.

The Change Request workflow now performs controlled state transitions and approval decisions.
Authorization must therefore be explicit, testable and independent from HTTP transport details.

The system is expected to introduce OIDC for authentication and Cedar for policy evaluation later.
The authorization model must allow that implementation to be introduced without coupling the
domain layer to an identity provider or policy engine.

## Decision

### 1. Separate authentication and authorization

Authentication establishes the actor identity (`subject`).

Authorization answers whether that actor may perform a specific action on a resource.

The domain layer does not depend on OIDC, LDAP, Cedar, HTTP headers or other identity infrastructure.

### 2. Application-layer authorization boundary

Authorization is enforced at the application/use-case boundary, not only in HTTP handlers.

HTTP middleware/handlers are responsible for obtaining the authenticated principal and passing it
to application use cases.

Until a trusted principal resolver is implemented, `X-Actor-ID` is available only through the explicit
`development-header` mode for local/manual workflow testing. Startup rejects this mode unless the
listener binds to a loopback IP address. With `ATLASIS_AUTH_MODE` unset, API routes fail closed with
HTTP 401. Roles and organization scope must never be accepted from untrusted request headers or
request bodies; they must come from verified identity claims or server-side mappings.

Application use cases are responsible for asking an authorization port before performing a protected operation.

The planned port is conceptually:

```go
type Authorizer interface {
    Authorize(
        ctx context.Context,
        subject Subject,
        action Action,
        resource Resource,
    ) error
}
```

The concrete port shape may evolve during implementation, but the contract must remain independent
of the future policy engine.

### 3. Default deny

An action is denied unless an explicit authorization rule grants it.

Authorization failures are represented by `ports.ErrForbidden` and are mapped by the HTTP layer to `403 Forbidden`.

### 4. Roles are permission bundles

Roles are not persisted in domain entities and are not used directly by domain rules.

An actor may have multiple roles.

Initial logical roles:

- Initiator
- Reviewer
- Approver
- Asset Owner
- Security Officer
- Administrator

The role-to-permission mapping is a policy concern and may later move from application code to Cedar policies.

### 5. Permissions are action-oriented

Initial permissions are explicit actions rather than endpoint names:

| Resource | Action |
|---|---|
| information_asset | read |
| information_asset | create |
| change_request | read |
| change_request | create |
| change_request | submit |
| change_request | review |
| change_request | request_changes |
| change_request | reject |
| change_request | approve |
| change_request | apply |
| approval | read |
| approval | create |
| approval | approve |
| approval | reject |

Endpoint names are transport details and must not become the authorization model.

### 6. Object-level authorization

Some permissions require resource attributes in addition to the actor's role.

The primary example is an approval decision:

`approval.approve` or `approval.reject` is allowed only when the actor is the
assigned approver for that approval.

This remains a domain invariant today and must also be represented in the future authorization policy.

The same pattern will be used for asset ownership and organization-scoped permissions.

### 7. Workflow transition rules remain in the domain

Authorization and workflow validity are different concerns.

Example:

- Authorization: may actor X request approval of CR00001?
- Domain: is CR00001 currently in a state from which APPROVED is a valid transition?

A successful authorization check must never bypass domain state-transition validation.

### 8. Administrator is not an implicit bypass

The Administrator role is not treated as an unconditional superuser in domain code.

Administrative permissions must be explicit policy grants.

Security-sensitive actions such as approval and application remain subject to normal authorization and domain rules unless a later ADR explicitly defines a break-glass mechanism.

## Initial role / permission matrix

| Role | Primary permissions |
|---|---|
| Initiator | information_asset.read, information_asset.create, change_request.read, change_request.create, change_request.submit, approval.read |
| Reviewer | information_asset.read, change_request.read, change_request.review, change_request.request_changes, change_request.reject, change_request.approve, approval.read, approval.create |
| Approver | approval.read, approval.approve, approval.reject |
| Asset Owner | information_asset.read, change_request.read, approval.read |
| Security Officer | information_asset.read, change_request.read, change_request.review, change_request.request_changes, approval.read, approval.create |
| Administrator | explicit administrative permissions defined by policy; no implicit workflow bypass |

The matrix is intentionally a starting policy model. Resource ownership, organization scope and
separation-of-duties constraints are evaluated as conditions, not encoded as additional roles.

## Separation of duties

The model supports independent actors for:

- initiating a Change Request;
- reviewing a Change Request;
- providing a required approval;
- applying an approved Change Request.

A later policy iteration may prohibit a single subject from occupying incompatible roles on the same Change Request.

## Initial executable RBAC policy

The first in-process policy adapter uses the role matrix above and denies access by default.
Every decision requires a known subject, known role set, known action/resource pairing, and an
explicit organization scope that contains the resource's organization.

The adapter also evaluates relationship attributes that must be sourced by the server:

- Initiators may read and submit only their own Change Requests.
- Asset Owners may read only assets and Change Requests for which they are the recorded owner.
- Approvers may read and decide only approvals assigned to them.
- Reviewer and Security Officer permissions are limited to organizations included in the subject's trusted scope.
- Administrator does not bypass these checks.

The policy adapter is a first step toward Cedar, not a substitute for authentication. It must be
called with a subject built from a verified principal and resource attributes loaded or derived by
the server. It must not be wired to roles supplied by a client.

The change_request.apply action is intentionally denied to all current roles until the responsible execution
role and its separation-of-duties requirements are explicitly decided. This avoids silently treating
Administrator or Reviewer as an unrestricted executor.

## Testing requirements

Authorization becomes a required application-level test dimension.

At minimum, protected use cases must test:

1. allowed actor/action/resource;
2. denied actor/action/resource returning `ports.ErrForbidden`;
3. object-level restriction for assigned approver;
4. authorization failure does not mutate resource state;
5. domain transition failure remains a conflict/invalid-transition result even when authorization succeeds.

## Future Cedar integration

Cedar will implement the authorization port rather than become a dependency of domain objects.

The intended flow is:

OIDC identity
  -> authenticated Principal
  -> application authorization request
  -> Cedar policy evaluation
  -> allow/deny
  -> domain/use-case execution

Replacing the current local/trusted identity mechanism with OIDC and replacing the authorization
implementation with Cedar should therefore not require changes to domain invariants or persistence.
