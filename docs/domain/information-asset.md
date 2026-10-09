# Legacy Information Asset Model

> **Status:** describes the current legacy implementation. The agreed target architecture separates Information Systems and Objects of Information Infrastructure into distinct domain entities. Do not extend this unified model for new target behavior without an explicit design decision.

## Legacy identity

- ID format: IA00000.
- The ID is sequential and does not encode asset type.
- Asset type is a separate attribute.
- Current types: information_system, information_infrastructure_object.

This identity model is retained for compatibility and migration mapping. It is not the target identifier design for new IS/OII entities.

## Legacy lifecycle status

- DRAFT
- ACTIVE
- SUSPENDED
- RETIRED

The valid transitions and semantics must be verified against current domain code and tests.

## Legacy core attributes

- id
- type
- name
- short_name
- status
- organization_id
- owner_id
- purpose
- criticality
- risk_level

## Legacy security profile

- protection_required
- protection_status
- attestation_status
- cyber_center_required

The completeness, meanings and validation rules for target IS/OII attributes must be established in a separate normative matrix tied to applicable Republic of Belarus requirements, including OAC orders No. 66 and No. 130.

## Current persistence and versioning

The current schema separates the legacy asset's current-version pointer from immutable snapshots:

    information_assets
        current_version
              |
              v
    information_asset_versions
        v1
        v2
        v3

A version is intended to be a complete snapshot of the legacy asset state. Controlled changes are performed through Change Requests.

## Current invariants to preserve during migration

1. Legacy IDs are stable and must be preserved as migration references.
2. Current version must point to an existing immutable version.
3. Applied CR and version-pointer changes are persisted atomically.
4. Every successful CR application creates a new version.
5. Concurrency checks prevent stale CRs from overwriting newer state.
6. Existing CRs, approvals and audit-relevant links must remain traceable.

## Target transition

The target model defines separate IS and OII entities with their own technical IDs, display numbers, attributes, ownership and versions. One IS may belong to multiple OIIs; nested OIIs are not supported in v1. Composition changes target the OII through a CR and are part of the OII version snapshot. Membership does not transfer ownership or permissions.

See:
- docs/architecture/README.md
- docs/architecture/data-model-v1.md
- docs/architecture/implementation-plan-v1.md
- docs/project-context.md
