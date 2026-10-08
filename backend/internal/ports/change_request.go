package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

type ChangeRequestRepository interface {
	Create(ctx context.Context, cr change.ChangeRequest) error
	GetByID(ctx context.Context, id change.ID) (*change.ChangeRequest, error)
	Save(ctx context.Context, cr change.ChangeRequest) error
}
