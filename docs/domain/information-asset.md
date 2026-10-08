# Information Asset

## Identity

- ID format: `IA00000`
- ID is sequential and does not encode asset type.
- Asset type is a separate attribute.

## Asset Types

- `information_system`
- `information_infrastructure_object`

## Lifecycle Status

- `DRAFT`
- `ACTIVE`
- `SUSPENDED`
- `RETIRED`

## Core Attributes

- `id`
- `type`
- `name`
- `short_name`
- `status`
- `organization_id`
- `owner_id`
- `purpose`
- `criticality`
- `risk_level`

## Security Profile

- `protection_required`
- `protection_status`
- `attestation_status`
- `cyber_center_required`

## Versioning

Every approved state is immutable.

```text
information_assets
    current_version
          |
          v
information_asset_versions
    v1
    v2
    v3
```

A version contains a complete snapshot of the asset state.

Controlled changes are performed only through Change Requests.

## Invariants

1. ID must match IA + exactly five digits.
2. ID cannot be changed after creation.
3. Organization and asset owner must exist.
4. Current version must always point to an existing immutable version.
5. Approved asset state cannot be modified directly.
6. Every approved change creates a new version.
7. Asset lifecycle transitions must be explicitly defined and validated.
