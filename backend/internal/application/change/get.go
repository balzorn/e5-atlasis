package change

import (
	"context"
	"fmt"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type GetChangeRequestUseCase struct {
	repository ports.ChangeRequestRepository
}

func NewGetChangeRequestUseCase(repository ports.ChangeRequestRepository) *GetChangeRequestUseCase {
	return &GetChangeRequestUseCase{repository: repository}
}

func (uc *GetChangeRequestUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	changeRequest, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get change request: %w", err)
	}

	if changeRequest == nil {
		return nil, fmt.Errorf("change request %q: %w", id, ports.ErrNotFound)
	}

	return changeRequest, nil
}
