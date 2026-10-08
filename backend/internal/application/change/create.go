package change

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const (
	maxChangeRequestInitiatorLength = 128
	maxChangeRequestTitleLength     = 500
	maxChangeRequestDescription     = 4000
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

	if cmd.BaseVersion < 1 {
		return nil, fmt.Errorf("%w: base version must be greater than zero", ports.ErrInvalidInput)
	}

	if current.CurrentVersion.Int() != cmd.BaseVersion {
		return nil, fmt.Errorf(
			"%w: base version %d does not match current asset version %d",
			ports.ErrConflict,
			cmd.BaseVersion,
			current.CurrentVersion.Int(),
		)
	}

	initiator := strings.TrimSpace(cmd.Initiator)
	if initiator == "" {
		return nil, fmt.Errorf("%w: initiator is required", ports.ErrInvalidInput)
	}
	if utf8.RuneCountInString(initiator) > maxChangeRequestInitiatorLength {
		return nil, fmt.Errorf("%w: initiator is too long", ports.ErrInvalidInput)
	}

	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ports.ErrInvalidInput)
	}
	if utf8.RuneCountInString(title) > maxChangeRequestTitleLength {
		return nil, fmt.Errorf("%w: title is too long", ports.ErrInvalidInput)
	}

	description := strings.TrimSpace(cmd.Description)
	if utf8.RuneCountInString(description) > maxChangeRequestDescription {
		return nil, fmt.Errorf("%w: description is too long", ports.ErrInvalidInput)
	}

	changes := make([]domainchange.FieldChange, 0, len(cmd.Changes))
	seenFields := make(map[string]struct{}, len(cmd.Changes))

	for i, proposal := range cmd.Changes {
		field := proposal.Field
		if field == "" {
			return nil, fmt.Errorf("%w: field is required", ports.ErrInvalidInput)
		}

		if _, exists := seenFields[string(field)]; exists {
			return nil, fmt.Errorf("%w: field %q appears more than once", ports.ErrInvalidInput, field)
		}
		seenFields[string(field)] = struct{}{}

		if err := proposal.NewValue.ValidateFor(field); err != nil {
			return nil, fmt.Errorf("%w: %v", ports.ErrInvalidInput, err)
		}

		oldValue, err := current.FieldValue(field)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ports.ErrInvalidInput, err)
		}

		if oldValue.Equal(proposal.NewValue) {
			return nil, fmt.Errorf("%w: field %q has no actual change", ports.ErrInvalidInput, field)
		}

		changes = append(changes, domainchange.FieldChange{
			ID:       fmt.Sprintf("CHG%05d", i+1),
			Field:    field,
			OldValue: oldValue,
			NewValue: proposal.NewValue,
		})
	}

	if len(changes) == 0 {
		return nil, fmt.Errorf("%w: at least one change is required", ports.ErrInvalidInput)
	}

	id, err := uc.ids.Next(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	cr := domainchange.ChangeRequest{
		ID:          id,
		AssetID:     cmd.AssetID.String(),
		BaseVersion: cmd.BaseVersion,
		Status:      domainchange.StatusDraft,
		Initiator:   initiator,
		Title:       title,
		Description: description,
		Changes:     changes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := cr.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrInvalidInput, err)
	}

	if err := uc.changeRequests.Create(ctx, cr); err != nil {
		return nil, err
	}

	return &cr, nil
}
