package ports

import (
	"context"

	"github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

type ChangeRequestIDGenerator interface {
	Next(ctx context.Context) (change.ID, error)
}
