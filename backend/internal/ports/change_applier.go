package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

// ChangeApplier atomically persists an applied Change Request,
// the new current Information Asset state and its new immutable version.
type ChangeApplier interface {
	Apply(
		ctx context.Context,
		cr change.ChangeRequest,
		asset asset.InformationAsset,
		version asset.AssetVersion,
	) error
}
