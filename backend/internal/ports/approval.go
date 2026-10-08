package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
)

type ApprovalRepository interface {
	Create(ctx context.Context, a approval.Approval) error
	GetByID(ctx context.Context, id approval.ID) (*approval.Approval, error)
	ListByChangeRequestID(
		ctx context.Context,
		changeRequestID string,
	) ([]approval.Approval, error)
	Save(ctx context.Context, a approval.Approval) error
}
