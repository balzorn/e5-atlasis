package approval

import (
	"context"
	"fmt"
	"time"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type DecisionCommand struct {
	ApprovalID domainapproval.ID
	DecidedBy  string
	Comment    string
}

type ApproveApprovalUseCase struct {
	repository ports.ApprovalRepository
}

func NewApproveApprovalUseCase(
	repository ports.ApprovalRepository,
) *ApproveApprovalUseCase {
	return &ApproveApprovalUseCase{
		repository: repository,
	}
}

func (uc *ApproveApprovalUseCase) Execute(
	ctx context.Context,
	cmd DecisionCommand,
) (*domainapproval.Approval, error) {
	return decide(ctx, uc.repository, cmd, true)
}

type RejectApprovalUseCase struct {
	repository ports.ApprovalRepository
}

func NewRejectApprovalUseCase(
	repository ports.ApprovalRepository,
) *RejectApprovalUseCase {
	return &RejectApprovalUseCase{
		repository: repository,
	}
}

func (uc *RejectApprovalUseCase) Execute(
	ctx context.Context,
	cmd DecisionCommand,
) (*domainapproval.Approval, error) {
	return decide(ctx, uc.repository, cmd, false)
}

func decide(
	ctx context.Context,
	repository ports.ApprovalRepository,
	cmd DecisionCommand,
	approve bool,
) (*domainapproval.Approval, error) {
	a, err := repository.GetByID(ctx, cmd.ApprovalID)
	if err != nil {
		return nil, err
	}

	if a == nil {
		return nil, fmt.Errorf(
			"approval %q not found",
			cmd.ApprovalID,
		)
	}

	now := time.Now().UTC()

	if approve {
		err = a.Approve(cmd.DecidedBy, cmd.Comment, now)
	} else {
		err = a.Reject(cmd.DecidedBy, cmd.Comment, now)
	}

	if err != nil {
		if a.Status != domainapproval.StatusPending {
			return nil, fmt.Errorf("%w: approval %q is no longer pending", ports.ErrConflict, a.ID)
		}
		if a.ApproverID != cmd.DecidedBy {
			return nil, fmt.Errorf("%w: user %q is not assigned as approver", ports.ErrForbidden, cmd.DecidedBy)
		}
		return nil, fmt.Errorf("%w: invalid approval decision", ports.ErrInvalidInput)
	}

	if err := repository.Save(ctx, *a); err != nil {
		return nil, err
	}

	return a, nil
}
