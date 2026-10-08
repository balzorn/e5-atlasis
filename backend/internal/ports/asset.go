package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

type AssetRepository interface {
	Create(
		ctx context.Context,
		asset asset.InformationAsset,
		version asset.AssetVersion,
	) error

	GetByID(ctx context.Context, id asset.AssetID) (*asset.InformationAsset, error)
}
