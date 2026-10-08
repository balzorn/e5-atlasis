package change

import (
	"context"
	"fmt"
	"time"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type StartReviewUseCase struct {
	changeRequests ports.ChangeRequestRepository
}

func NewStartReviewUseCase(
	changeRequests ports.ChangeRequestRepository,
) *StartReviewUseCase {
	return &StartReviewUseCase{
		changeRequests: changeRequests,
	}
}

func (uc *StartReviewUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return transitionChangeRequest(
		ctx,
		uc.changeRequests,
		id,
		domainchange.StatusUnderReview,
	)
}

type RequestChangesUseCase struct {
	changeRequests ports.ChangeRequestRepository
}

func NewRequestChangesUseCase(
	changeRequests ports.ChangeRequestRepository,
) *RequestChangesUseCase {
	return &RequestChangesUseCase{
		changeRequests: changeRequests,
	}
}

func (uc *RequestChangesUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return transitionChangeRequest(
		ctx,
		uc.changeRequests,
		id,
		domainchange.StatusChangesRequested,
	)
}

type RejectChangeRequestUseCase struct {
	changeRequests ports.ChangeRequestRepository
}

func NewRejectChangeRequestUseCase(
	changeRequests ports.ChangeRequestRepository,
) *RejectChangeRequestUseCase {
	return &RejectChangeRequestUseCase{
		changeRequests: changeRequests,
	}
}

func (uc *RejectChangeRequestUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return transitionChangeRequest(
		ctx,
		uc.changeRequests,
		id,
		domainchange.StatusRejected,
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

	if err := cr.TransitionTo(target); err != nil {
		return nil, err
	}

	cr.UpdatedAt = time.Now().UTC()

	if err := repository.Save(ctx, *cr); err != nil {
		return nil, err
	}

	return cr, nil
}
