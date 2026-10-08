package asset

import (
	"context"
	"fmt"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type GetAssetUseCase struct {
	repository ports.AssetRepository
}

func NewGetAssetUseCase(repository ports.AssetRepository) *GetAssetUseCase {
	return &GetAssetUseCase{repository: repository}
}

func (uc *GetAssetUseCase) Execute(
	ctx context.Context,
	id domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	asset, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get information asset: %w", err)
	}

	if asset == nil {
		return nil, fmt.Errorf("information asset %q: %w", id, ports.ErrNotFound)
	}

	return asset, nil
}
