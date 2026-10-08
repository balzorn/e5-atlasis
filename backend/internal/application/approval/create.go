package approval

import (
	"context"
	"fmt"
	"strings"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const (
	maxApprovalApproverIDLength = 128
	maxApprovalCommentLength    = 4000
)

type CreateApprovalCommand struct {
	ChangeRequestID domainchange.ID
	Type            domainapproval.Type
	Required        bool
	ApproverID      string
}

type CreateApprovalUseCase struct {
	changeRequests ports.ChangeRequestRepository
	approvals      ports.ApprovalRepository
	ids             ports.ApprovalIDGenerator
}

func NewCreateApprovalUseCase(
	changeRequests ports.ChangeRequestRepository,
	approvals ports.ApprovalRepository,
	ids ports.ApprovalIDGenerator,
) *CreateApprovalUseCase {
	return &CreateApprovalUseCase{
		changeRequests: changeRequests,
		approvals:      approvals,
		ids:             ids,
	}
}

func (uc *CreateApprovalUseCase) Execute(
	ctx context.Context,
	cmd CreateApprovalCommand,
) (*domainapproval.Approval, error) {
	cr, err := uc.changeRequests.GetByID(ctx, cmd.ChangeRequestID)
	if err != nil {
		return nil, err
	}

	if cr == nil {
		return nil, fmt.Errorf("change request %q not found", cmd.ChangeRequestID)
	}

	if cr.Status != domainchange.StatusUnderReview {
		return nil, fmt.Errorf(
			"%w: approvals can only be created for change requests UNDER_REVIEW",
			ports.ErrConflict,
		)
	}

	approverID := strings.TrimSpace(cmd.ApproverID)
	if approverID == "" || len(approverID) > maxApprovalApproverIDLength {
		return nil, fmt.Errorf("%w: invalid approver ID", ports.ErrInvalidInput)
	}

	switch cmd.Type {
	case domainapproval.TypeAssetOwner,
		domainapproval.TypeSecurity,
		domainapproval.TypeOrganization,
		domainapproval.TypeSystemOwner:
	default:
		return nil, fmt.Errorf("%w: invalid approval type", ports.ErrInvalidInput)
	}

	id, err := uc.ids.Next(ctx)
	if err != nil {
		return nil, err
	}

	a := domainapproval.Approval{
		ID:              id,
		ChangeRequestID: cr.ID.String(),
		Type:            cmd.Type,
		Status:          domainapproval.StatusPending,
		Required:        cmd.Required,
		ApproverID:      approverID,
	}

	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("validate approval: %w", err)
	}

	if err := uc.approvals.Create(ctx, a); err != nil {
		return nil, err
	}

	return &a, nil
}

type ListApprovalsUseCase struct {
	changeRequests ports.ChangeRequestRepository
	approvals      ports.ApprovalRepository
}

func NewListApprovalsUseCase(
	changeRequests ports.ChangeRequestRepository,
	approvals ports.ApprovalRepository,
) *ListApprovalsUseCase {
	return &ListApprovalsUseCase{
		changeRequests: changeRequests,
		approvals:      approvals,
	}
}

func (uc *ListApprovalsUseCase) Execute(
	ctx context.Context,
	id domainchange.ID,
) ([]domainapproval.Approval, error) {
	cr, err := uc.changeRequests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cr == nil {
		return nil, fmt.Errorf("change request %q not found", id)
	}

	return uc.approvals.ListByChangeRequestID(ctx, id.String())
}
