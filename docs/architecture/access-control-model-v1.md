# AtlasIS Access Control Model v1

- Status: Proposed
- Version: 1.0-draft
- Related decision: ADR-011
- Date: 2026-10-09

## 1. Purpose and scope

This document describes the proposed authorization model for AtlasIS v1. It is a design specification
for review, not a statement that the model has already been implemented.

The model supports multiple independent tenants in one AtlasIS deployment, a catalogue of
organizations and typed relationships within each tenant, scoped role assignments, delegated
participant management, auditable access administration, and policy-controlled Change Request
workflow. The initial deployment may use one application and PostgreSQL database; future separate
deployments must remain possible without changing authorization semantics.

This document does not authorize changes to code, database migrations, identity-provider integration,
Cedar policies or deployment topology. Those changes follow only after the documentation diff is
approved.

## 2. Core concepts

### 2.1 Tenant

A tenant is the top-level isolation boundary for a registry customer, holding or independently
administered organization/group. The current holding is represented as one tenant. A separate,
unrelated legal entity can be represented as another tenant without joining the holding's organization
catalogue.

Each tenant has its own:
- registry data and organization catalogue;
- role assignments and delegation boundaries;
- policy context and policy configuration;
- audit history and access-administration history.

Tenant identifiers are stable internal identifiers. Display names, legal names, domains and organization
names are not tenant identifiers and must not be used as authorization proof.

### 2.2 Organization

An organization is an independent entity inside exactly one tenant. It may represent a legal entity
or another managed organizational unit as the domain model evolves. Organizations have stable
identifiers and lifecycle status.

The tenant's organization catalogue may store typed relationships between organizations, for example:
- member of group;
- managed by;
- provides service to;
- successor of;
- another explicitly registered relationship type.

The relationship catalogue describes context and does not merge organizations. A relationship does
not automatically grant access, transfer ownership, inherit assignments or make an organization a
child authorization scope. Policies may explicitly evaluate particular relationship types, but that
behavior must be documented and tested.

### 2.3 Registry object

A registry object is an information system, object of informatization, Change Request, approval,
access assignment, audit event or another managed record.

Every tenant-owned object must be attributable to one tenant. Objects that are organization-owned
must also reference one organization belonging to that same tenant. Workflow records must retain
tenant context from their associated resource and must not accept a client-supplied tenant identifier
as authoritative.

### 2.4 Subject

A subject is an authenticated actor represented by a stable identity identifier. Roles, organization
membership, tenant membership and delegated authority are obtained from trusted identity claims or
server-side records. They are not trusted merely because they appear in request headers or bodies.

### 2.5 Role and role assignment

A role is a named bundle of possible permissions. A role assignment grants a subject a role within
an explicit tenant and scope, subject to validity, delegation and policy conditions.

A role name alone never proves authorization. Every decision also considers tenant, scope, action,
resource attributes and relevant relationships.

### 2.6 Technical identifiers and human-readable numbers

Every persisted entity has two distinct identifiers with different purposes:

- `id` (technical identifier): proposed UUIDv7, used for primary keys, foreign-key relationships,
  API resource identity and machine-to-machine references.
- `display_number` (human-readable number): a prefixed incrementing number used in screens,
  discussions, documents and support conversations. It is not a database relationship key and is
  not proof of authorization.

The human-readable number is unique only within its declared numbering scope. Initial rules:

| Entity | Prefix example | Numbering scope | Example |
|---|---|---|---|
| Organization | `ORG` | Within one tenant | `ORG00007` |
| Information Asset (IA) | `IA` | Within its owning organization | `IA00001` |
| Change Request | `CR` | Within the Information Asset it concerns | `CR00001` |
| Approval | `APR` | Within its Change Request | `APR00001` |
| Thread | `THR` | Within the resource being discussed | `THR00001` |
| Comment | `CMT` | Within its thread; if no thread exists, within its direct parent resource | `CMT00001` |

Consequently, two organizations in the same tenant may each have an `IA00001`; two information
assets may each have a `CR00001`; and two Change Requests may each have an `APR00001`. This is
intentional: each sequence is local to its entity type and parent. Creating records under one tenant,
organization, asset or request must not advance another scope's sequence.

A short number alone may be ambiguous outside its parent context. User-facing references, copied
links and support instructions must include the parent path needed to identify the record uniquely,
for example `ORG00007-IA00001-CR00001` and
`ORG00007-IA00001-CR00001-THR00001-CMT00001`. The precise separators and whether the full path is
shown everywhere can be settled in UI/API design; ambiguity must not be introduced into machine
interfaces. Internally, child-to-parent relationships always use technical IDs.

UUIDv7 is proposed because its time component can improve locality for ordered database indexes
compared with random UUIDv4. It does not guarantee strict chronological ordering across concurrent
writers and does not remove the need for primary-key and foreign-key constraints. Technical IDs
must not encode tenant membership as a substitute for an explicit tenant boundary.

#### Counter generation and integrity

Counter allocation must be safe under concurrent requests. The implementation must not use
`COUNT(*) + 1`, an unlocked read-modify-write, or a process-local counter. Use a durable
database-backed counter keyed by tenant, entity type and parent scope as applicable, with atomic
allocation/locking and a uniqueness constraint on the scope plus display number. Number allocation
and record creation should be transactionally coordinated. Gaps caused by rollback, deletion or
failed attempts are acceptable; issued display numbers must not be reused.

Before implementation, define the exact counter schema, uniqueness constraints, deletion/archive
semantics, and backfill strategy for existing data. All creation paths (API, jobs, imports and future
integrations) must use the same allocation service. Tests must cover concurrent creation, same-parent
sequence increments, independent counters across parents and tenants, rollback, and attempted
duplicate allocation.

## 3. Tenant isolation requirements

1. Every authorization request has a trusted tenant context.
2. The subject must have an active membership or other explicit authorization basis for that tenant.
3. The target resource must belong to the same tenant context.
4. A request cannot change its effective tenant by changing a body field, query parameter or header.
5. Cross-tenant reads and writes are denied by default.
6. Repository queries and writes must be tenant-aware; filtering only in the user interface or HTTP
   handler is insufficient.
7. List, search, count, export, autocomplete, nested-resource and error paths must not leak data from
   other tenants.
8. Tenant isolation must be tested before enabling multi-tenant use.
9. A shared application/database is an initial deployment choice, not a permanent domain assumption.
10. Separate deployment of a tenant in the future must not require changing the meaning of roles,
    actions or policy decisions.

## 4. Scope model

The proposed scope types are:

| Scope type | Meaning | Default effect |
|---|---|---|
| tenant | Entire tenant | Only explicitly allowed tenant-wide functions |
| organization | One organization in a tenant | Only resources and actions explicitly granted for that organization |
| information_system | One information system | Only that system and explicitly supported related actions |
| change_request | One Change Request | Only that workflow object and explicitly supported actions |
| approval | One approval record | Only that approval and explicitly supported actions |

An assignment scope is not automatically inherited by related objects. For example, organization
scope does not automatically grant every action on every information system of that organization
unless the relevant policy explicitly defines that relationship.

The organization catalogue's relationships are not authorization scopes. In particular, the system
must not treat “member of group” as permission to access every other organization in that group.

A tenant-wide role assignment must name its tenant scope explicitly. Platform-wide roles, if needed,
are a separate administration concept and must not be represented as ordinary tenant-wide roles.

## 5. Role catalogue

| Role | Purpose | Constraints |
|---|---|---|
| Initiator | Initiate and submit changes | Limited by assignment scope and request/resource relationships |
| Reviewer | Review change requests | No implicit approval or apply rights |
| Approver | Decide approvals | Must be assigned to the approval; self-approval depends on change policy |
| Asset Owner | Manage assigned information systems | Participant management limited by delegation |
| IS Participant | Participate in an assigned information system | Only explicitly granted object actions |
| Security Officer | Perform information-security oversight | Organization-scoped by default; tenant-wide authority assigned separately |
| Access Administrator | Administer access assignments | Only within an explicitly delegated administration scope |
| Platform Administrator | Administer platform operation | No implicit access to tenant business data or workflow bypass |
| Change Executor | Apply approved changes | Requires explicit grant and successful policy/workflow checks |
| Auditor | Read permitted audit history | No mutation permissions |

The catalogue is logical. The final permission mapping must be implemented as explicit policy, not
inferred from role labels or job titles.

## 6. Role assignment record

The proposed logical assignment record contains:

| Attribute | Required | Meaning |
|---|---|---|
| assignment_id | Yes | Stable unique identifier |
| tenant_id | Yes | Tenant in which the assignment is valid |
| subject_id | Yes | Subject receiving the role |
| role | Yes | Role identifier |
| scope_type | Yes | Scope category |
| scope_id | Yes | Identifier of tenant, organization or supported object scope |
| granted_by | Yes | Actor who created the assignment |
| justification | Yes | Business/security reason for the assignment |
| status | Yes | Pending, active, revoked or expired as appropriate to the workflow |
| valid_from | Yes | Start of validity |
| valid_until | No | Expiry time, if bounded |
| created_at | Yes | Record creation timestamp |
| updated_at | Yes | Last record update timestamp |

This is a conceptual schema, not a database migration specification. The exact status lifecycle,
identity foreign keys, uniqueness constraints and storage types are to be finalized in the data-design
stage.

Every assignment must be validated against the tenant and scope catalogue. A subject cannot receive
an assignment in a tenant or organization for which the subject lacks an authorized membership or
administrative basis. Revocation and expiry must take effect in authorization decisions without
requiring a role name to be removed from unrelated scopes.

## 7. Delegated management of IS participants

An Asset Owner may manage participants for an information system only if:
- the owner has an active, trusted assignment for that system or a policy-recognized ownership relation;
- participant management is explicitly granted for the relevant action;
- the target participant and system belong to the same tenant;
- the owner is not granting a role, scope or duration beyond their delegation;
- any required approval or separation-of-duties policy is satisfied.

Delegation must be bounded. An owner cannot create tenant-wide assignments, grant platform-level
authority, expand their own scope, or administer unrelated systems merely because they own one IS.

Participant-management operations include listing participants, granting or revoking a participant
assignment, changing scope, changing validity, and changing the participant's permitted role.
The final action vocabulary may distinguish these operations more finely.

Each operation must generate an audit event recording the actor, target subject, tenant, system,
operation, old and new effective assignment where applicable, justification, time and outcome.
The assignment mutation and required audit record must be committed atomically or through a
durable transactional mechanism.

## 8. Action catalogue

Initial action vocabulary:

| Resource | Actions |
|---|---|
| information_asset | read, create, update, archive |
| information_system | read, create, update, archive, read_participants, manage_participants |
| change_request | read, create, update_draft, submit, review, request_changes, reject, approve, apply |
| approval | read, create, approve, reject |
| access_assignment | read, grant, revoke, change_scope |
| audit_event | read, export |
| policy | read, propose_change, approve_change, publish |

Actions are application capabilities, not HTTP routes. Implementation may split an action where
different risk levels or audit requirements justify it. New actions default to denied until a policy
grant and tests are supplied.

## 9. Policy evaluation model

A decision evaluates the following inputs as applicable:

1. trusted tenant context;
2. authenticated subject identity;
3. active role assignments and their validity windows;
4. action and resource type;
5. resource tenant and organization;
6. assignment scope and delegation limits;
7. object relationships, such as recorded owner or assigned approver;
8. resource/workflow state and change type;
9. separation-of-duties conditions;
10. explicit policy rules and default deny.

The policy implementation may begin as in-process RBAC with contextual conditions and later move
to Cedar. Cedar is a policy engine, not a source of truth for identity, tenant membership, ownership,
or assignment lifecycle. Those facts must be supplied by trusted server-side systems.

Policy evaluation and domain workflow validation remain separate. An allow decision permits an
attempt; it does not make an invalid state transition valid.

## 10. Separation of duties and Change Executor

Applying a change is a distinct capability assigned to Change Executor. No other role receives this
capability implicitly.

Self-approval rules depend on change type and applicable policy. The model does not impose a single
universal rule that all self-approval is always forbidden or always allowed. For each change type,
the policy must specify the allowed combinations of initiator, reviewer, approver and executor and
the required number or type of independent decisions.

If a policy prohibits a combination, it must be checked using trusted actor identities and the
workflow's recorded history, not client-provided values. Domain validation must enforce the required
workflow sequence independently of authorization.

## 11. Audit and observability

Audit events should contain, as applicable:
- event_id and occurred_at;
- tenant_id;
- event_type;
- actor_id and target_subject_id;
- action;
- resource_type and resource_id;
- decision and outcome;
- reason_code;
- policy_id and policy_version;
- assignment_id or related delegation reference;
- justification;
- request_id and trace_id/span_id where available;
- source/system identifier.

At minimum, audit access-assignment grants, revocations, scope changes, expiry changes, participant
management, authorization denials for sensitive operations, and policy lifecycle events.

Audit records are durable security evidence. OpenTelemetry is for traces, metrics, logs and
correlation; it does not replace authoritative audit storage. Reliable forwarding to the cybersecurity
center or another consumer should be designed separately, with a transactional outbox considered
where delivery must not be lost.

## 12. Error handling and information disclosure

Authorization denials should map to a consistent forbidden response for authenticated requests.
Unauthenticated requests fail authentication. Cross-tenant resource lookups should not disclose
whether another tenant's identifier exists; the external response policy must be consistent and
documented.

Authorization failures must not mutate domain state. Audit may record the denial, subject to the
defined audit policy, without exposing sensitive internal policy details to the caller.

## 13. Required test matrix

| Test area | Minimum cases |
|---|---|
| Tenant boundary | Same-tenant allow; cross-tenant read/write denied |
| Information disclosure | Lists, searches, counts, exports and nested lookups are tenant-filtered |
| Organization scope | Same-org allow; other-org deny absent explicit assignment |
| Organization relationships | Relationship alone grants no access |
| Assignment validity | Not-yet-valid, expired and revoked assignments deny |
| Delegation | Owner can manage within bounds; cannot grant broader role/scope |
| Assignment audit | Grant, revoke, scope change and expiry change produce durable audit |
| Approval | Assigned approver can decide; unassigned actor cannot |
| Self-approval | Results match policy for each change type |
| Change execution | Only authorized Change Executor may apply; invalid workflow state still fails |
| Default deny | Unknown role/action/resource or missing context denies |
| Mutation safety | Denied operation leaves domain state unchanged |
| Policy migration | RBAC and Cedar decisions are equivalent over the agreed fixture set |

## 14. Implementation sequence after documentation approval

1. Approve ADR-011 and this specification.
2. Inventory existing data and define the initial tenant and organization mapping.
3. Finalize domain terminology, identifiers, organization relationship types and lifecycle rules, including UUIDv7 technical IDs, scoped display-number counters and parent-qualified references.
4. Finalize authorization request/decision contract and trusted principal/assignment resolution.
5. Define persistence changes, constraints, tenant-aware repository behavior and migration/rollback plan.
6. Implement application-level authorization in use cases; remove direct trust in raw actor headers.
7. Add tenant-isolation, organization-scope, delegation, separation-of-duties and audit tests.
8. Implement durable audit capture and define outbox/consumer behavior if external delivery is required.
9. Define Cedar schema/policies and decision-equivalence tests before switching policy engines.
10. Roll out multi-tenant behavior only after cross-tenant isolation tests pass.

No code, migration, Cedar policy or runtime configuration change is included in this documentation proposal.
