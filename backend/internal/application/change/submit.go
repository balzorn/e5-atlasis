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
	return &SubmitChangeRequestUseCase{changeRequests: changeRequests}
}

func (uc *SubmitChangeRequestUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return transitionChangeRequest(
		ctx,
		uc.changeRequests,
		id,
		domainchange.StatusSubmitted,
	)
}

func transitionChangeRequest(
	ctx context.Context,
	repository ports.ChangeRequestRepository,
	id domainchange.ID,
	target domainchange.Status,
) (*domainchange.ChangeRequest, error) {
	cr, err := repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cr == nil {
		return nil, fmt.Errorf("change request %q not found", id)
	}

	expectedStatus := cr.Status
	if err := cr.TransitionTo(target); err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrConflict, err)
	}

	cr.UpdatedAt = time.Now().UTC()

	if err := repository.Save(ctx, *cr, expectedStatus); err != nil {
		return nil, err
	}

	return cr, nil
}
