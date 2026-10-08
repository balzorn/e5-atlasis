package asset

import (
	"context"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type CreateAssetCommand struct {
	Type           domainasset.AssetType
	Name           string
	ShortName      string
	OrganizationID string
	OwnerID        string
	Purpose        string
	Criticality    domainasset.Criticality
	RiskLevel      domainasset.RiskLevel
	Security       domainasset.SecurityProfile
}

type CreateAssetUseCase struct {
	repository ports.AssetRepository
	ids        ports.AssetIDGenerator
}

func NewCreateAssetUseCase(
	repository ports.AssetRepository,
	ids ports.AssetIDGenerator,
) *CreateAssetUseCase {
	return &CreateAssetUseCase{
		repository: repository,
		ids:        ids,
	}
}

func (uc *CreateAssetUseCase) Execute(
	ctx context.Context,
	cmd CreateAssetCommand,
) (*domainasset.InformationAsset, error) {
	id, err := uc.ids.Next(ctx)
	if err != nil {
		return nil, err
	}

	a := domainasset.InformationAsset{
		ID:             id,
		Type:           cmd.Type,
		Name:           cmd.Name,
		ShortName:      cmd.ShortName,
		Status:         domainasset.AssetStatusDraft,
		OrganizationID: cmd.OrganizationID,
		OwnerID:        cmd.OwnerID,
		Purpose:        cmd.Purpose,
		Criticality:    cmd.Criticality,
		RiskLevel:      cmd.RiskLevel,
		Security:       cmd.Security,
		CurrentVersion: 1,
	}

	if err := a.Validate(); err != nil {
		return nil, err
	}

	version := domainasset.AssetVersion{
		AssetID: a.ID,
		Version: 1,
		State:   a,
	}

	if err := uc.repository.Create(ctx, a, version); err != nil {
		return nil, err
	}

	return &a, nil
}
