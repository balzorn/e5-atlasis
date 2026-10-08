package change

import (
	"context"
	"fmt"
	"time"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type ApproveChangeRequestUseCase struct {
	changeRequests ports.ChangeRequestRepository
	approvals      ports.ApprovalRepository
}

func NewApproveChangeRequestUseCase(
	changeRequests ports.ChangeRequestRepository,
	approvals ports.ApprovalRepository,
) *ApproveChangeRequestUseCase {
	return &ApproveChangeRequestUseCase{
		changeRequests: changeRequests,
		approvals:      approvals,
	}
}

func (uc *ApproveChangeRequestUseCase) Execute(
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

	approvals, err := uc.approvals.ListByChangeRequestID(ctx, cr.ID.String())
	if err != nil {
		return nil, err
	}

	if len(approvals) == 0 {
		return nil, fmt.Errorf("%w: change request %q has no approvals", ports.ErrConflict, cr.ID)
	}

	for _, a := range approvals {
		if a.Required && a.Status != domainapproval.StatusApproved {
			return nil, fmt.Errorf(
				"%w: required approval %q is not approved",
				ports.ErrConflict,
				a.ID,
			)
		}
	}

	expectedStatus := cr.Status
	if err := cr.TransitionTo(domainchange.StatusApproved); err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrConflict, err)
	}

	cr.UpdatedAt = time.Now().UTC()

	if err := uc.changeRequests.Save(ctx, *cr, expectedStatus); err != nil {
		return nil, err
	}

	return cr, nil
}
