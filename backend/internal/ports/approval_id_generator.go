package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
)

type ApprovalIDGenerator interface {
	Next(ctx context.Context) (approval.ID, error)
}
