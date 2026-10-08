package asset

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	CreatedBy      string
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
	a := domainasset.InformationAsset{
		Type:           cmd.Type,
		Name:           strings.TrimSpace(cmd.Name),
		ShortName:      strings.TrimSpace(cmd.ShortName),
		Status:         domainasset.AssetStatusDraft,
		OrganizationID: strings.TrimSpace(cmd.OrganizationID),
		OwnerID:        strings.TrimSpace(cmd.OwnerID),
		Purpose:        strings.TrimSpace(cmd.Purpose),
		Criticality:    cmd.Criticality,
		RiskLevel:      cmd.RiskLevel,
		Security:       cmd.Security,
		CurrentVersion: 1,
	}

	if err := a.ValidateAttributes(); err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrInvalidInput, err)
	}

	createdBy := strings.TrimSpace(cmd.CreatedBy)
	if createdBy == "" {
		return nil, fmt.Errorf("%w: creator is required", ports.ErrInvalidInput)
	}
	if len(createdBy) > 128 {
		return nil, fmt.Errorf("%w: creator is too long", ports.ErrInvalidInput)
	}

	id, err := uc.ids.Next(ctx)
	if err != nil {
		return nil, err
	}

	a.ID = id
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("validate generated asset: %w", err)
	}

	version := domainasset.AssetVersion{
		AssetID:   a.ID,
		Version:   1,
		State:     a,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
	}

	if err := uc.repository.Create(ctx, a, version); err != nil {
		return nil, err
	}

	return &a, nil
}
