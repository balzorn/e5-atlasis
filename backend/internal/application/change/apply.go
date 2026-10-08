package change

import (
	"context"
	"fmt"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type ApplyChangeRequestUseCase struct {
	assets  ports.AssetRepository
	changes ports.ChangeRequestRepository
	applier ports.ChangeApplier
}

func NewApplyChangeRequestUseCase(
	assets ports.AssetRepository,
	changes ports.ChangeRequestRepository,
	applier ports.ChangeApplier,
) *ApplyChangeRequestUseCase {
	return &ApplyChangeRequestUseCase{
		assets:  assets,
		changes: changes,
		applier: applier,
	}
}

func (uc *ApplyChangeRequestUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainasset.InformationAsset, error) {
	cr, err := uc.changes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cr == nil {
		return nil, fmt.Errorf("change request %q not found", id)
	}

	if cr.Status != domainchange.StatusApproved {
		return nil, fmt.Errorf(
			"%w: change request %q must be APPROVED, got %s",
			ports.ErrConflict,
			id,
			cr.Status,
		)
	}

	assetID, err := domainasset.ParseAssetID(cr.AssetID)
	if err != nil {
		return nil, err
	}

	current, err := uc.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}

	if current == nil {
		return nil, fmt.Errorf("asset %q not found", assetID)
	}

	if current.CurrentVersion.Int() != cr.BaseVersion {
		return nil, fmt.Errorf(
			"%w: asset version %d does not match change request base version %d",
			ports.ErrConflict,
			current.CurrentVersion.Int(),
			cr.BaseVersion,
		)
	}

	next := *current

	for _, fieldChange := range cr.Changes {
		if err := next.ApplyField(
			fieldChange.Field,
			fieldChange.NewValue,
		); err != nil {
			return nil, fmt.Errorf(
				"apply field %q: %w",
				fieldChange.Field,
				err,
			)
		}
	}

	nextVersion, err := domainasset.NewVersion(
		current.CurrentVersion.Int() + 1,
	)
	if err != nil {
		return nil, err
	}

	next.CurrentVersion = nextVersion

	now := time.Now().UTC()

	version := domainasset.AssetVersion{
		ID: fmt.Sprintf(
			"%s-v%05d",
			assetID,
			nextVersion.Int(),
		),
		AssetID:         assetID,
		Version:         nextVersion,
		State:           next,
		CreatedBy:       cr.Initiator,
		CreatedAt:       now,
		ChangeRequestID: stringPointer(cr.ID.String()),
	}

	if err := next.Validate(); err != nil {
		return nil, err
	}

	if err := cr.TransitionTo(domainchange.StatusApplying); err != nil {
		return nil, err
	}

	if err := cr.TransitionTo(domainchange.StatusApplied); err != nil {
		return nil, err
	}

	cr.UpdatedAt = now

	if err := uc.applier.Apply(
		ctx,
		*cr,
		next,
		version,
	); err != nil {
		return nil, err
	}

	return &next, nil
}

func stringPointer(value string) *string {
	return &value
}
