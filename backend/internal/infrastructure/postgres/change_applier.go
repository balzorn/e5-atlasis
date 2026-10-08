package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type ChangeApplier struct {
	db *DB
}

func NewChangeApplier(db *DB) *ChangeApplier {
	return &ChangeApplier{db: db}
}

func (r *ChangeApplier) Apply(
	ctx context.Context,
	cr domainchange.ChangeRequest,
	a domainasset.InformationAsset,
	version domainasset.AssetVersion,
) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin change application transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentVersion int

	err = tx.QueryRow(
		ctx,
		`
		SELECT current_version
		FROM information_assets
		WHERE id = $1
		FOR UPDATE
		`,
		a.ID.String(),
	).Scan(&currentVersion)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("asset %q not found", a.ID)
		}

		return fmt.Errorf("lock information asset: %w", err)
	}

	if currentVersion != cr.BaseVersion {
		return fmt.Errorf(
			"asset %q version conflict: current=%d, expected=%d",
			a.ID,
			currentVersion,
			cr.BaseVersion,
		)
	}

	if version.Version.Int() != currentVersion+1 {
		return fmt.Errorf(
			"invalid next version: got=%d, expected=%d",
			version.Version.Int(),
			currentVersion+1,
		)
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

	result, err := tx.Exec(
		ctx,
		`
		UPDATE information_assets
		SET current_version = $2
		WHERE id = $1
		  AND current_version = $3
		`,
		a.ID.String(),
		version.Version.Int(),
		cr.BaseVersion,
	)
	if err != nil {
		return fmt.Errorf("update current asset version: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf(
			"asset %q version changed during application",
			a.ID,
		)
	}

	result, err = tx.Exec(
		ctx,
		`
		UPDATE change_requests
		SET
			status = $2,
			updated_at = $3
		WHERE id = $1
		  AND status = 'APPROVED'
		`,
		cr.ID.String(),
		cr.Status,
		cr.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update change request status: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf(
			"change request %q is not in APPROVED state",
			cr.ID,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit change application: %w", err)
	}

	return nil
}

var _ interface {
	Apply(
		context.Context,
		domainchange.ChangeRequest,
		domainasset.InformationAsset,
		domainasset.AssetVersion,
	) error
} = (*ChangeApplier)(nil)

var _ ports.ChangeApplier = (*ChangeApplier)(nil)
