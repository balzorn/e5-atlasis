package postgres

import (
	"context"
	"fmt"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type AssetRepository struct {
	db *DB
}

func NewAssetRepository(db *DB) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) Create(
	ctx context.Context,
	a domainasset.InformationAsset,
	version domainasset.AssetVersion,
) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin asset creation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO information_assets (
			id,
			current_version
		)
		VALUES ($1, $2)
		`,
		a.ID.String(),
		a.CurrentVersion.Int(),
	)
	if err != nil {
		return fmt.Errorf("insert information asset: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO information_asset_versions (
			asset_id,
			version,
			type,
			name,
			short_name,
			status,
			organization_id,
			owner_id,
			purpose,
			criticality,
			risk_level,
			protection_required,
			protection_status,
			attestation_status,
			cyber_center_required,
			created_by,
			created_at,
			change_request_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		)
		`,
		version.AssetID.String(),
		version.Version.Int(),
		version.State.Type,
		version.State.Name,
		version.State.ShortName,
		version.State.Status,
		version.State.OrganizationID,
		version.State.OwnerID,
		version.State.Purpose,
		version.State.Criticality,
		version.State.RiskLevel,
		version.State.Security.ProtectionRequired,
		version.State.Security.ProtectionStatus,
		version.State.Security.AttestationStatus,
		version.State.Security.CyberCenterRequired,
		version.CreatedBy,
		version.CreatedAt,
		version.ChangeRequestID,
	)
	if err != nil {
		return fmt.Errorf("insert asset version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit asset creation: %w", err)
	}

	return nil
}

func (r *AssetRepository) GetByID(
	ctx context.Context,
	id domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	var (
		currentVersion      int
		version             int
		assetType           domainasset.AssetType
		name                string
		shortName           string
		status              domainasset.AssetStatus
		organizationID      string
		ownerID             string
		purpose             string
		criticality         domainasset.Criticality
		riskLevel           domainasset.RiskLevel
		protectionRequired  bool
		protectionStatus    domainasset.ProtectionStatus
		attestationStatus   domainasset.AttestationStatus
		cyberCenterRequired bool
	)

	err := r.db.pool.QueryRow(
		ctx,
		`
		SELECT
			a.current_version,
			v.version,
			v.type,
			v.name,
			v.short_name,
			v.status,
			v.organization_id,
			v.owner_id,
			v.purpose,
			v.criticality,
			v.risk_level,
			v.protection_required,
			v.protection_status,
			v.attestation_status,
			v.cyber_center_required
		FROM information_assets a
		JOIN information_asset_versions v
		  ON v.asset_id = a.id
		 AND v.version = a.current_version
		WHERE a.id = $1
		`,
		id.String(),
	).Scan(
		&currentVersion,
		&version,
		&assetType,
		&name,
		&shortName,
		&status,
		&organizationID,
		&ownerID,
		&purpose,
		&criticality,
		&riskLevel,
		&protectionRequired,
		&protectionStatus,
		&attestationStatus,
		&cyberCenterRequired,
	)
	if err != nil {
		return nil, fmt.Errorf("get information asset: %w", err)
	}

	return &domainasset.InformationAsset{
		ID:             id,
		Type:           assetType,
		Name:           name,
		ShortName:      shortName,
		Status:         status,
		OrganizationID: organizationID,
		OwnerID:        ownerID,
		Purpose:        purpose,
		Criticality:    criticality,
		RiskLevel:      riskLevel,
		Security: domainasset.SecurityProfile{
			ProtectionRequired:  protectionRequired,
			ProtectionStatus:    protectionStatus,
			AttestationStatus:   attestationStatus,
			CyberCenterRequired: cyberCenterRequired,
		},
		CurrentVersion: domainasset.Version(currentVersion),
	}, nil
}

var _ ports.AssetRepository = (*AssetRepository)(nil)
