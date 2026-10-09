# ADR-011: Authorization Model

- Status: Proposed
- Date: 2026-10-09

## Context

The current API reads the caller identifier from the `X-Actor-ID` request header.
This is a temporary development placeholder, not authentication: an arbitrary client can spoof it
unless a trusted, authenticated gateway removes any client-supplied header and injects a verified
actor identifier. The API must not rely on this header as a production security boundary.

The Change Request workflow performs controlled state transitions and approval decisions.
Authorization must therefore be explicit, testable and independent from HTTP transport details.

AtlasIS is intended to support more than one independently administered customer or legal entity
within a shared application deployment. The current holding is one tenant, not a special system-wide
authorization scope. A tenant contains a catalogue of organizations and explicit relationships
between them. Organizations remain independent entities; a relationship such as “managed by” or
“member of group” does not itself grant access.

Keycloak is the selected OIDC provider for v1, and verified OIDC claims are the initial boundary
through which AtlasIS receives user identity/profile attributes from the corporate Active Directory
environment. Cedar may be introduced for policy evaluation without coupling the domain layer to an
identity provider or policy engine. The deployment may start with one shared application and
PostgreSQL database, while preserving a path to separate deployments for different tenants.

## Decision

### 1. Separate authentication and authorization

Authentication establishes the actor identity (subject). Authorization answers whether that actor
may perform a specific action on a resource in a specific tenant and scope.

The domain layer does not depend on OIDC, LDAP, Cedar, HTTP headers or other identity infrastructure.
Keycloak/OIDC is an infrastructure adapter; verified identity claims are mapped to an internal AtlasIS
subject before application use cases run.

### 1a. Internal subject and external identity mapping

AtlasIS assigns each person/subject an immutable internal UUID. External identity is represented
separately from that internal key.

For v1:
- Keycloak is the OIDC provider. AtlasIS reads profile attributes only from verified claims.
- Corporate Active Directory is the initial authoritative directory for user identity attributes.
  A future corporate identity service with its own API may replace or supplement this source.
- Map the validated OIDC issuer/subject pair (iss, sub) to the internal subject. Retain sAMAccountName
  and other AD/Keycloak attributes as namespaced external attributes or lookup keys, not as the
  internal primary key.
- Do not assume sAMAccountName, email, UPN, domain name or display name is globally unique or immutable.
  Store the identity source/namespace and define uniqueness within that namespace. A change in email,
  domain or organization must not create a new internal subject if the same external identity remains.
- The first OIDC login may provision a minimal subject record only under an explicit account-linking
  and tenant-assignment policy. Successful authentication does not itself grant access to registry data.
- AtlasIS owns role assignments, authorization scopes, membership decisions and their audit history.
  Administrators or delegated access administrators must be able to assign roles to a previously
  authenticated subject after first login. Claims may supply profile attributes but must not be treated
  as an unreviewed source of business permissions.
- Moving a person between organizations requires an explicit update/review of AtlasIS organization
  membership and scoped role assignments. Future identity-source migration must support a controlled,
  audited link from a new external identity to an existing internal subject.

This mapping keeps authorization stable when mutable directory attributes change and avoids coupling
the data model to one directory implementation.

### 2. Tenant is the data-isolation boundary

Every tenant-owned record and every authorization decision must be associated with exactly one
tenant. Tenant identity must be established from trusted server-side context, not accepted from an
untrusted request body or header as proof of access.

A tenant is the top-level boundary for:
- registry data and organization catalogue;
- role assignments and delegated authority;
- authorization policy configuration;
- audit events and access-administration history.

The initial deployment may host multiple tenants in one application and PostgreSQL database.
Application and persistence boundaries must nevertheless enforce tenant isolation. The design must
not assume that all tenants will always share one physical deployment or database.

A tenant is not the same thing as an organization, holding, or legal entity. The current holding
is represented as a tenant with its own organization catalogue. Other unrelated organizations or
groups may be represented in separate tenants.

### 3. Organizations are independent entities with explicit relationships

Each organization belongs to one tenant. A tenant maintains a catalogue of organizations and may
record typed relationships between them, such as “member of group”, “managed by”, or another
approved relationship type.

Organization relationships describe organizational context only. They do not automatically:
- merge organization identities;
- transfer ownership of information systems;
- inherit role assignments;
- grant read or write access;
- expand the scope of an organization-scoped Security Officer.

Any relationship-based authorization must be an explicit, testable policy decision. There is no
implicit access inheritance through the organization catalogue.

### 4. Application-layer authorization boundary

Authorization is enforced at the application/use-case boundary, not only in HTTP handlers.
HTTP middleware and handlers obtain the authenticated principal and pass it to application use cases.

Until a trusted principal resolver is implemented, X-Actor-ID is available only through the explicit
development-header mode for local/manual workflow testing. Startup rejects this mode unless the
listener binds to a loopback IP address. With ATLASIS_AUTH_MODE unset, API routes fail closed with
HTTP 401. Roles, tenant membership and organization scope must never be accepted from untrusted
request headers or request bodies; they must come from verified identity claims or server-side
mappings.

Application use cases ask an authorization port before performing a protected operation. The
authorization request includes the trusted tenant context, subject, action and resource context.
The concrete port shape may evolve, but the contract remains independent of the future policy engine.

### 5. Default deny and explicit scope

An action is denied unless an explicit authorization rule grants it. Authorization failures are
represented by ports.ErrForbidden and mapped by the HTTP layer to 403 Forbidden.

Role assignments are scoped. Supported conceptual scope types are:
- tenant — tenant-wide authority, only for explicitly designated functions;
- organization — one organization within a tenant;
- information_system or other registered object — an individual object, where the role supports it;
- change_request or approval — a specific workflow object, where the role supports it.

The tenant is not represented as a special organization scope. An organization-scoped assignment
does not grant tenant-wide access. An assignment to an organization or object does not automatically
expand to related organizations or child objects unless a specific policy explicitly grants that
behavior.

### 6. Roles are permission bundles; assignments are contextual

Roles are not persisted as domain attributes on information-system records and are not used directly
as domain workflow rules. A subject may hold multiple role assignments with different scopes.

Initial logical roles:
- Initiator
- Reviewer
- Approver
- Asset Owner
- IS Participant
- Security Officer
- Access Administrator
- Platform Administrator
- Change Executor
- Auditor

Role assignment records must identify at least the subject, role, tenant, scope type and identifier,
granting actor, justification, status, start time and optional expiry. Assignment and revocation
must be audited.

Global roles are allowed only for functions that explicitly require tenant-wide or platform-wide
authority. Platform-wide authority is distinct from tenant-wide authority and must be separately
defined; it must not arise from an ordinary tenant role.

### 7. Delegation and management of IS participants

An IS Owner may manage participants of that information system only within the owner's delegated
authority. The owner cannot grant a role, scope or duration that exceeds the owner's own delegation
or the policy's delegation limits. The owner cannot use participant management to grant tenant-wide
or unrelated-organization access.

All participant and role assignment changes must be audited, including grant, revoke, scope change,
expiry change, actor, target subject, affected resource, justification and authorization outcome.
A successful write must not be committed without its required audit record being durably captured.

### 8. Initial permission actions

Permissions are explicit actions rather than endpoint names. The initial action vocabulary includes:

| Resource | Actions |
|---|---|
| information_system | read, create, update, archive, read_participants, manage_participants |
| informatization_object | read, create, update, archive, manage_composition |
| change_request | read, create, update_draft, submit, review, request_changes, reject, approve, apply |
| approval | read, create, approve, reject |
| access_assignment | read, grant, revoke, change_scope |
| audit_event | read, export |
| policy | read, propose_change, approve_change, publish |

The vocabulary is a design contract and may be refined during implementation. Endpoint names remain
transport details and must not become the authorization model.

### 9. Organization scope and Security Officer

Security Officer authority is organization-scoped by default. A Security Officer assigned to one
organization does not automatically receive authority across the tenant or holding.

Tenant-wide Security Officer authority requires a separate, explicit assignment with a tenant scope.
The same principle applies to other organization-based roles. A broader role must not be inferred
from job title, organizational relationship or membership in a group.

### 10. Object-level authorization and workflow rules

Some permissions require resource attributes in addition to role and scope. Examples include:
- Initiators may read or submit only Change Requests they are entitled to access;
- Asset Owners may manage participants only for information systems they own and only within delegated limits;
- Approvers may decide only approvals assigned to them;
- organization-scoped roles are limited to resources in the assigned organization, subject to explicit policy;
- every resource and subject in a decision must belong to the same trusted tenant context.

Authorization and workflow validity are different concerns. Authorization answers whether an actor may
attempt an action; domain logic determines whether the resource is in a state where the transition
is valid. A successful authorization check never bypasses domain state-transition validation.

### 11. Separation of duties and Change Executor

Applying an approved Change Request is a separate permission held by the Change Executor role.
Administrator, Reviewer, Approver and Security Officer do not receive apply permission implicitly.
The apply action remains denied unless the Change Executor grant and required policy conditions are met.

Self-approval is not a universal hard-coded rule. Whether the initiator, reviewer, approver or executor
may be the same person depends on change type and the applicable policy. These constraints must be
expressed explicitly and tested for each relevant change type. Domain workflow validation remains
mandatory even when authorization succeeds.

Administrator is not an unconditional superuser. Administrative permissions are explicit policy
grants. Security-sensitive actions remain subject to authorization, separation-of-duties and domain
rules unless a future ADR defines a controlled break-glass mechanism.

### 12. Audit and observability

Security-relevant access decisions and access-administration changes must be auditable. Audit events
should capture event identifier, timestamp, tenant, actor, action, resource type and identifier,
decision/outcome, reason code, policy identifier/version where available, request correlation
identifier and trace/span identifiers where available.

Audit storage is authoritative for audit history. OpenTelemetry provides telemetry and correlation,
not a substitute for durable audit records. Reliable publication to external consumers, including a
cybersecurity center, may use a transactional outbox and is specified separately from authorization
decision semantics.

### 13. Initial executable RBAC and future Cedar integration

The initial in-process authorization adapter uses explicit grants and denies by default. It must
validate the tenant context, known subject and role assignments, action/resource compatibility,
resource organization and object-level relationships. The adapter must be invoked with a subject
and assignments obtained from trusted server-side sources, never client-supplied role or scope data.

Cedar will implement the authorization port rather than become a dependency of domain objects.
Tenant and organization attributes, assignment scope, ownership, participant relationships, assigned
approver, change type/status and separation-of-duties conditions must be available to policy evaluation
through trusted server-side data.

Replacing the local/trusted identity mechanism with OIDC and replacing the authorization
implementation with Cedar should not require changes to domain invariants or workflow persistence.

## Initial role / permission intent

| Role | Initial intent and limits |
|---|---|
| Initiator | Create and submit Change Requests within assigned scope; access own or otherwise authorized requests |
| Reviewer | Review requests within assigned scope; no implicit apply permission |
| Approver | Decide assigned approvals, subject to change-type policy and separation of duties |
| Asset Owner | Read owned information systems and manage their participants within delegated limits |
| IS Participant | Access explicitly assigned information systems and permitted actions |
| Security Officer | Security oversight within assigned organization by default; tenant-wide authority requires a separate assignment |
| Access Administrator | Manage role assignments only within explicitly delegated administration scope |
| Platform Administrator | Operate platform-level functions; no implicit access to tenant business data or workflow bypass |
| Change Executor | Apply approved changes when explicitly authorized and all policy/workflow conditions are met |
| Auditor | Read permitted audit history; no mutation permissions |

This table describes role intent, not an unconditional grant list. Concrete grants must be expressed
as explicit policy and scoped role assignments.

### 14. Stable technical IDs and scoped display numbers

Every persisted entity uses a stable technical identifier, proposed as UUIDv7, for database
relationships, API resource identity and machine-to-machine references. UUIDv7 is not a secret and
does not replace authentication or authorization. Its time-ordered layout is an implementation
benefit, not a guarantee of strict global ordering or uniqueness without database constraints.

Human-readable numbers are a separate presentation/reference attribute. They use a type prefix and
an incrementing sequence whose scope is explicitly defined per entity type. No tenant's activity may
advance another tenant's counter, and creating one kind of child entity must not advance a counter
for another kind or parent.

Initial numbering rules:

- Organization numbers are allocated within a tenant.
- Information System (IS) and Object of Informatization (OII) numbers use separate prefixes and
  counters scoped to the tenant. They do not change when the owning organization changes.
- Change Request (CR) numbers are allocated within the single target resource (one IS or one OII).
  Each target resource starts its own CR sequence. An OII composition change targets the OII.
- Individually managed non-IS technical assets and their TA numbering are deferred from v1.
- Approval, thread and comment numbers are allocated within their defined parent: approvals within a
  Change Request; threads within the resource they discuss; comments within their thread (or directly
  within the parent resource if the model has no thread).
- Other child entities must define their parent and counter scope explicitly before implementation;
  they must not silently use a global sequence.

Because a short local number is not necessarily unique by itself, user-facing references and copied
links must include enough parent context to resolve unambiguously, for example
`IS00001-CR00001` or `OII00001-CR00001-THR00001-CMT00001`. The exact separator
and display format may be finalized in the UI/API design, but the scope semantics above are fixed by
this decision. Legacy IA identifiers must be mapped and retained as references during migration; they
are not the target numbering scheme for newly created IS/OII records. Internally, relationships always use technical IDs, never display numbers.

Counter allocation must be concurrency-safe and transactional. Implementations must not derive the
next number using `COUNT(*) + 1` or an unlocked read-modify-write. Use a database-backed counter
keyed by tenant, entity type and parent scope (as applicable), protected by atomic update/locking,
and enforce a uniqueness constraint on the resulting scope plus display number. Gaps after failed
transactions or deleted records are acceptable; reusing an issued number is not. Tests must cover
parallel creation, rollback behavior, same-parent sequencing and isolation between tenants and
parents.

This is an architectural proposal for identifier semantics, not a migration or code change. Existing
records, if any, require an explicit backfill and compatibility plan before implementation.

## Testing requirements

Authorization is a required application-level test dimension. At minimum, tests must cover:
1. allowed and denied actor/action/resource combinations;
2. cross-tenant access denial for reads and writes;
3. no data leakage through list, search, counts, errors or related-resource lookups across tenants;
4. organization-scoped denial when a resource belongs to another organization in the same tenant;
5. no implicit access through an organization relationship;
6. owner delegation limits for participant and role management;
7. assignment grant, revoke, scope change and expiry audit;
8. Change Executor-only application and policy-dependent self-approval restrictions;
9. authorization failure does not mutate resource state;
10. domain transition failure remains a conflict/invalid-transition result even when authorization succeeds.

## Migration and rollout

Before implementation, inventory existing records and define how they are assigned to an initial tenant
and organization. Existing role or organization assumptions must not be converted into tenant-wide
permissions implicitly. Tenant identity must be propagated through repository queries and writes,
and tenant isolation must be tested before multi-tenant use is enabled.

This ADR records the architecture. The companion access-model specification defines the proposed
scope and assignment semantics, action catalogue, audit requirements and implementation sequence.
