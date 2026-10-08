package change

import (
	"context"
	"fmt"
	"time"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type SubmitChangeRequestUseCase struct {
	changeRequests ports.ChangeRequestRepository
}

func NewSubmitChangeRequestUseCase(
	changeRequests ports.ChangeRequestRepository,
) *SubmitChangeRequestUseCase {
	return &SubmitChangeRequestUseCase{
		changeRequests: changeRequests,
	}
}

func (uc *SubmitChangeRequestUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	cr, err := uc.changeRequests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cr == nil {
		return nil, fmt.Errorf("change request %q not found", id)
	}

	if err := cr.TransitionTo(domainchange.StatusSubmitted); err != nil {
		return nil, err
	}

	cr.UpdatedAt = time.Now().UTC()

	if err := uc.changeRequests.Save(ctx, *cr); err != nil {
		return nil, err
	}

	return cr, nil
}
