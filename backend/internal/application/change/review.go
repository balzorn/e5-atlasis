package change

import (
	"context"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type StartReviewUseCase struct {
	changeRequests ports.ChangeRequestRepository
}

func NewStartReviewUseCase(
	changeRequests ports.ChangeRequestRepository,
) *StartReviewUseCase {
	return &StartReviewUseCase{changeRequests: changeRequests}
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
	return &RequestChangesUseCase{changeRequests: changeRequests}
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
	return &RejectChangeRequestUseCase{changeRequests: changeRequests}
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
