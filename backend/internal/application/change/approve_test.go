package change

import (
	"context"
	"testing"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

type fakeApprovalRepository struct {
	approvals []domainapproval.Approval
}

func (r *fakeApprovalRepository) Create(
	_ context.Context,
	a domainapproval.Approval,
) error {
	r.approvals = append(r.approvals, a)
	return nil
}

func (r *fakeApprovalRepository) GetByID(
	_ context.Context,
	id domainapproval.ID,
) (*domainapproval.Approval, error) {
	for _, a := range r.approvals {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, nil
}

func (r *fakeApprovalRepository) ListByChangeRequestID(
	_ context.Context,
	changeRequestID string,
) ([]domainapproval.Approval, error) {
	return r.approvals, nil
}

func (r *fakeApprovalRepository) Save(
	_ context.Context,
	a domainapproval.Approval,
) error {
	for i := range r.approvals {
		if r.approvals[i].ID == a.ID {
			r.approvals[i] = a
			return nil
		}
	}
	r.approvals = append(r.approvals, a)
	return nil
}

func newApprovalTestChangeRepository(
	status domainchange.Status,
) *fakeChangeRequestRepository {
	return &fakeChangeRequestRepository{
		created: &domainchange.ChangeRequest{
			ID:        "CR00001",
			AssetID:   "IA00001",
			Status:    status,
			Initiator: "USR001",
			Title:     "Test change",
			Changes: []domainchange.FieldChange{
				{
					ID:       "CHG00001",
					Field:    "owner_id",
					OldValue: "USR001",
					NewValue: "USR002",
				},
			},
		},
	}
}

func TestApproveChangeRequestWithRequiredApproval(t *testing.T) {
	crRepo := newApprovalTestChangeRepository(domainchange.StatusUnderReview)

	approvalRepo := &fakeApprovalRepository{
		approvals: []domainapproval.Approval{
			{
				ID:              "APR00001",
				ChangeRequestID: "CR00001",
				Type:            domainapproval.TypeSecurity,
				Status:          domainapproval.StatusApproved,
				Required:        true,
				ApproverID:      "USR002",
			},
		},
	}

	uc := NewApproveChangeRequestUseCase(crRepo, approvalRepo)

	cr, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.Status != domainchange.StatusApproved {
		t.Fatalf("Status = %s, want APPROVED", cr.Status)
	}
}

func TestApproveChangeRequestRejectsPendingRequiredApproval(t *testing.T) {
	crRepo := newApprovalTestChangeRepository(domainchange.StatusUnderReview)

	approvalRepo := &fakeApprovalRepository{
		approvals: []domainapproval.Approval{
			{
				ID:              "APR00001",
				ChangeRequestID: "CR00001",
				Type:            domainapproval.TypeSecurity,
				Status:          domainapproval.StatusPending,
				Required:        true,
				ApproverID:      "USR002",
			},
		},
	}

	uc := NewApproveChangeRequestUseCase(crRepo, approvalRepo)

	_, err := uc.Execute(context.Background(), "CR00001")
	if err == nil {
		t.Fatal("Execute() error = nil, want pending approval error")
	}
}

func TestApproveChangeRequestRejectsWithoutApprovals(t *testing.T) {
	crRepo := newApprovalTestChangeRepository(domainchange.StatusUnderReview)

	approvalRepo := &fakeApprovalRepository{}

	uc := NewApproveChangeRequestUseCase(crRepo, approvalRepo)

	_, err := uc.Execute(context.Background(), "CR00001")
	if err == nil {
		t.Fatal("Execute() error = nil, want no approvals error")
	}
}

func TestApproveChangeRequestAllowsPendingOptionalApproval(t *testing.T) {
	crRepo := newApprovalTestChangeRepository(domainchange.StatusUnderReview)

	approvalRepo := &fakeApprovalRepository{
		approvals: []domainapproval.Approval{
			{
				ID:              "APR00001",
				ChangeRequestID: "CR00001",
				Type:            domainapproval.TypeSecurity,
				Status:          domainapproval.StatusApproved,
				Required:        true,
				ApproverID:      "USR002",
			},
			{
				ID:              "APR00002",
				ChangeRequestID: "CR00001",
				Type:            domainapproval.TypeSystemOwner,
				Status:          domainapproval.StatusPending,
				Required:        false,
				ApproverID:      "USR003",
			},
		},
	}

	uc := NewApproveChangeRequestUseCase(crRepo, approvalRepo)

	cr, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.Status != domainchange.StatusApproved {
		t.Fatalf("Status = %s, want APPROVED", cr.Status)
	}
}
