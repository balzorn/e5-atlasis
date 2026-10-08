package change

import (
	"context"
	"fmt"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type CreateChangeRequestCommand struct {
	AssetID     domainasset.AssetID
	BaseVersion int
	Initiator   string
	Title       string
	Description string
	Changes     []ChangeProposal
}

type ChangeProposal struct {
	Field    domainasset.FieldName
	NewValue domainasset.FieldValue
}

type CreateChangeRequestUseCase struct {
	assets         ports.AssetRepository
	changeRequests ports.ChangeRequestRepository
	ids            ports.ChangeRequestIDGenerator
}

func NewCreateChangeRequestUseCase(
	assets ports.AssetRepository,
	changeRequests ports.ChangeRequestRepository,
	ids ports.ChangeRequestIDGenerator,
) *CreateChangeRequestUseCase {
	return &CreateChangeRequestUseCase{
		assets:         assets,
		changeRequests: changeRequests,
		ids:            ids,
	}
}

func (uc *CreateChangeRequestUseCase) Execute(
	ctx context.Context,
	cmd CreateChangeRequestCommand,
) (*domainchange.ChangeRequest, error) {
	current, err := uc.assets.GetByID(ctx, cmd.AssetID)
	if err != nil {
		return nil, err
	}

	if current.CurrentVersion.Int() != cmd.BaseVersion {
		return nil, fmt.Errorf(
			"base version %d does not match current asset version %d",
			cmd.BaseVersion,
			current.CurrentVersion.Int(),
		)
	}

	id, err := uc.ids.Next(ctx)
	if err != nil {
		return nil, err
	}

	changes := make([]domainchange.FieldChange, 0, len(cmd.Changes))
	seenFields := make(map[string]struct{}, len(cmd.Changes))

	for i, proposal := range cmd.Changes {
		field := proposal.Field

		if _, exists := seenFields[string(field)]; exists {
			return nil, fmt.Errorf("field %q appears more than once", field)
		}
		seenFields[string(field)] = struct{}{}

		if err := proposal.NewValue.ValidateFor(field); err != nil {
			return nil, err
		}

		oldValue, err := current.FieldValue(field)
		if err != nil {
			return nil, err
		}

		if oldValue.Equal(proposal.NewValue) {
			return nil, fmt.Errorf("field %q has no actual change", field)
		}

		changes = append(changes, domainchange.FieldChange{
			ID:       fmt.Sprintf("CHG%05d", i+1),
			Field:    field,
			OldValue: oldValue,
			NewValue: proposal.NewValue,
		})
	}

	now := time.Now().UTC()

	cr := domainchange.ChangeRequest{
		ID:          id,
		AssetID:     cmd.AssetID.String(),
		BaseVersion: cmd.BaseVersion,
		Status:      domainchange.StatusDraft,
		Initiator:   cmd.Initiator,
		Title:       cmd.Title,
		Description: cmd.Description,
		Changes:     changes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := cr.Validate(); err != nil {
		return nil, err
	}

	if err := uc.changeRequests.Create(ctx, cr); err != nil {
		return nil, err
	}

	return &cr, nil
}
