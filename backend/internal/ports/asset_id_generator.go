package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

type AssetIDGenerator interface {
	Next(ctx context.Context) (asset.AssetID, error)
}
