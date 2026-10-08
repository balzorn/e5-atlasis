package approval

import (
	"context"
	"testing"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
)

type fakeApprovalRepository struct {
	approval *domainapproval.Approval
	saved    *domainapproval.Approval
}

func (r *fakeApprovalRepository) Create(
	_ context.Context,
	a domainapproval.Approval,
) error {
	r.approval = &a
	return nil
}

func (r *fakeApprovalRepository) GetByID(
	_ context.Context,
	_ domainapproval.ID,
) (*domainapproval.Approval, error) {
	return r.approval, nil
}

func (r *fakeApprovalRepository) ListByChangeRequestID(
	_ context.Context,
	_ string,
) ([]domainapproval.Approval, error) {
	if r.approval == nil {
		return nil, nil
	}
	return []domainapproval.Approval{*r.approval}, nil
}

func (r *fakeApprovalRepository) Save(
	_ context.Context,
	a domainapproval.Approval,
) error {
	r.saved = &a
	r.approval = &a
	return nil
}

func TestApproveApproval(t *testing.T) {
	repo := &fakeApprovalRepository{
		approval: &domainapproval.Approval{
			ID:              "APR00001",
			ChangeRequestID: "CR00001",
			Type:            domainapproval.TypeSecurity,
			Status:          domainapproval.StatusPending,
			Required:        true,
			ApproverID:      "USR001",
		},
	}

	uc := NewApproveApprovalUseCase(repo)

	got, err := uc.Execute(context.Background(), DecisionCommand{
		ApprovalID: "APR00001",
		DecidedBy:  "USR001",
		Comment:    "Approved",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Status != domainapproval.StatusApproved {
		t.Fatalf("Status = %s, want APPROVED", got.Status)
	}

	if got.DecidedAt == nil {
		t.Fatal("DecidedAt is nil")
	}

	if repo.saved == nil {
		t.Fatal("approval was not saved")
	}
}

func TestRejectApproval(t *testing.T) {
	repo := &fakeApprovalRepository{
		approval: &domainapproval.Approval{
			ID:              "APR00001",
			ChangeRequestID: "CR00001",
			Type:            domainapproval.TypeSecurity,
			Status:          domainapproval.StatusPending,
			Required:        true,
			ApproverID:      "USR001",
		},
	}

	uc := NewRejectApprovalUseCase(repo)

	got, err := uc.Execute(context.Background(), DecisionCommand{
		ApprovalID: "APR00001",
		DecidedBy:  "USR001",
		Comment:    "Rejected",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Status != domainapproval.StatusRejected {
		t.Fatalf("Status = %s, want REJECTED", got.Status)
	}
}

func TestApprovalRejectsWrongApprover(t *testing.T) {
	repo := &fakeApprovalRepository{
		approval: &domainapproval.Approval{
			ID:              "APR00001",
			ChangeRequestID: "CR00001",
			Type:            domainapproval.TypeSecurity,
			Status:          domainapproval.StatusPending,
			Required:        true,
			ApproverID:      "USR001",
		},
	}

	uc := NewApproveApprovalUseCase(repo)

	_, err := uc.Execute(context.Background(), DecisionCommand{
		ApprovalID: "APR00001",
		DecidedBy:  "USR999",
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want wrong approver error")
	}
}
