CREATE TABLE information_assets (
    id TEXT PRIMARY KEY,
    current_version INTEGER NOT NULL DEFAULT 1,

    CONSTRAINT information_assets_id_chk
        CHECK (id ~ '^IA[0-9]{5}$'),

    CONSTRAINT information_assets_current_version_chk
        CHECK (current_version >= 1)
);

CREATE TABLE change_requests (
    id TEXT PRIMARY KEY,
    asset_id TEXT NOT NULL,
    base_version INTEGER NOT NULL,
    status TEXT NOT NULL,
    initiator TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT change_requests_asset_fk
        FOREIGN KEY (asset_id)
        REFERENCES information_assets(id),

    CONSTRAINT change_requests_base_version_chk
        CHECK (base_version >= 1),

    CONSTRAINT change_requests_status_chk
        CHECK (
            status IN (
                'DRAFT',
                'SUBMITTED',
                'UNDER_REVIEW',
                'CHANGES_REQUESTED',
                'APPROVED',
                'APPLYING',
                'APPLIED',
                'REJECTED',
                'CANCELLED'
            )
        )
);

CREATE TABLE information_asset_versions (
    asset_id TEXT NOT NULL,
    version INTEGER NOT NULL,

    type TEXT NOT NULL,
    name TEXT NOT NULL,
    short_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    organization_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    purpose TEXT NOT NULL DEFAULT '',
    criticality TEXT NOT NULL,
    risk_level TEXT NOT NULL,

    protection_required BOOLEAN NOT NULL,
    protection_status TEXT NOT NULL,
    attestation_status TEXT NOT NULL,
    cyber_center_required BOOLEAN NOT NULL,

    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    change_request_id TEXT,

    PRIMARY KEY (asset_id, version),

    CONSTRAINT asset_versions_asset_fk
        FOREIGN KEY (asset_id)
        REFERENCES information_assets(id),

    CONSTRAINT asset_versions_version_chk
        CHECK (version >= 1),

    CONSTRAINT asset_versions_type_chk
        CHECK (
            type IN (
                'information_system',
                'information_infrastructure_object'
            )
        ),

    CONSTRAINT asset_versions_status_chk
        CHECK (
            status IN (
                'DRAFT',
                'ACTIVE',
                'SUSPENDED',
                'RETIRED'
            )
        ),

    CONSTRAINT asset_versions_criticality_chk
        CHECK (
            criticality IN (
                'LOW',
                'MEDIUM',
                'HIGH',
                'CRITICAL'
            )
        ),

    CONSTRAINT asset_versions_risk_level_chk
        CHECK (
            risk_level IN (
                'LOW',
                'MEDIUM',
                'HIGH',
                'CRITICAL'
            )
        ),

    CONSTRAINT asset_versions_protection_status_chk
        CHECK (
            protection_status IN (
                'NOT_REQUIRED',
                'REQUIRED',
                'IN_PROGRESS',
                'IMPLEMENTED'
            )
        ),

    CONSTRAINT asset_versions_attestation_status_chk
        CHECK (
            attestation_status IN (
                'NOT_REQUIRED',
                'REQUIRED',
                'IN_PROGRESS',
                'ATTESTED',
                'EXPIRED'
            )
        )
);

ALTER TABLE information_assets
    ADD CONSTRAINT information_assets_current_version_fk
    FOREIGN KEY (id, current_version)
    REFERENCES information_asset_versions(asset_id, version)
    DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE change_requests
    ADD CONSTRAINT change_requests_base_version_fk
    FOREIGN KEY (asset_id, base_version)
    REFERENCES information_asset_versions(asset_id, version);

ALTER TABLE information_asset_versions
    ADD CONSTRAINT asset_versions_change_request_fk
    FOREIGN KEY (change_request_id)
    REFERENCES change_requests(id);

CREATE UNIQUE INDEX information_asset_versions_change_request_uq
    ON information_asset_versions(change_request_id)
    WHERE change_request_id IS NOT NULL;

CREATE TABLE change_request_changes (
    change_request_id TEXT NOT NULL,
    id TEXT NOT NULL,

    field TEXT NOT NULL,
    old_value JSONB NOT NULL,
    new_value JSONB NOT NULL,

    PRIMARY KEY (change_request_id, id),

    CONSTRAINT change_request_changes_cr_fk
        FOREIGN KEY (change_request_id)
        REFERENCES change_requests(id)
        ON DELETE CASCADE,

    CONSTRAINT change_request_changes_old_value_chk
        CHECK (jsonb_typeof(old_value) = 'object'),

    CONSTRAINT change_request_changes_new_value_chk
        CHECK (jsonb_typeof(new_value) = 'object'),

    CONSTRAINT change_request_changes_different_values_chk
        CHECK (old_value <> new_value)
);

CREATE TABLE approvals (
    id TEXT PRIMARY KEY,
    change_request_id TEXT NOT NULL,

    type TEXT NOT NULL,
    status TEXT NOT NULL,
    required BOOLEAN NOT NULL,

    approver_id TEXT NOT NULL,
    decided_at TIMESTAMPTZ,
    comment TEXT NOT NULL DEFAULT '',

    CONSTRAINT approvals_change_request_fk
        FOREIGN KEY (change_request_id)
        REFERENCES change_requests(id)
        ON DELETE CASCADE,

    CONSTRAINT approvals_type_chk
        CHECK (
            type IN (
                'ASSET_OWNER',
                'SECURITY',
                'ORGANIZATION',
                'SYSTEM_OWNER'
            )
        ),

    CONSTRAINT approvals_status_chk
        CHECK (
            status IN (
                'PENDING',
                'APPROVED',
                'REJECTED'
            )
        )
);

CREATE INDEX change_requests_asset_idx
    ON change_requests(asset_id);

CREATE INDEX change_requests_status_idx
    ON change_requests(status);

CREATE INDEX approvals_change_request_idx
    ON approvals(change_request_id);
