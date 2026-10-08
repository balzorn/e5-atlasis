package change

import (
	"context"
	"fmt"

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
	Field    string
	NewValue any
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

	for i, proposal := range cmd.Changes {
		oldValue, err := currentFieldValue(*current, proposal.Field)
		if err != nil {
			return nil, err
		}

		changes = append(changes, domainchange.FieldChange{
			ID:       fmt.Sprintf("CHG%05d", i+1),
			Field:    proposal.Field,
			OldValue: oldValue,
			NewValue: proposal.NewValue,
		})
	}

	cr := domainchange.ChangeRequest{
		ID:          id,
		AssetID:     cmd.AssetID.String(),
		BaseVersion: cmd.BaseVersion,
		Status:      domainchange.StatusDraft,
		Initiator:   cmd.Initiator,
		Title:       cmd.Title,
		Description: cmd.Description,
		Changes:     changes,
	}

	if err := cr.Validate(); err != nil {
		return nil, err
	}

	if err := uc.changeRequests.Create(ctx, cr); err != nil {
		return nil, err
	}

	return &cr, nil
}

func currentFieldValue(a domainasset.InformationAsset, field string) (any, error) {
	switch field {
	case "type":
		return a.Type, nil
	case "name":
		return a.Name, nil
	case "short_name":
		return a.ShortName, nil
	case "status":
		return a.Status, nil
	case "organization_id":
		return a.OrganizationID, nil
	case "owner_id":
		return a.OwnerID, nil
	case "purpose":
		return a.Purpose, nil
	case "criticality":
		return a.Criticality, nil
	case "risk_level":
		return a.RiskLevel, nil
	case "protection_required":
		return a.Security.ProtectionRequired, nil
	case "protection_status":
		return a.Security.ProtectionStatus, nil
	case "attestation_status":
		return a.Security.AttestationStatus, nil
	case "cyber_center_required":
		return a.Security.CyberCenterRequired, nil
	default:
		return nil, fmt.Errorf("field %q cannot be changed through a change request", field)
	}
}
