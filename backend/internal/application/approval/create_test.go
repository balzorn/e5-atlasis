package approval

import (
	"context"
	"errors"
	"testing"

	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type fakeChangeRequestRepository struct {
	changeRequest *domainchange.ChangeRequest
}

func (r *fakeChangeRequestRepository) Create(context.Context, domainchange.ChangeRequest) error {
	return nil
}

func (r *fakeChangeRequestRepository) GetByID(
	context.Context,
	domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return r.changeRequest, nil
}

func (r *fakeChangeRequestRepository) Save(
	context.Context,
	domainchange.ChangeRequest,
	domainchange.Status,
) error {
	return nil
}

type fakeApprovalIDGenerator struct {
	id domainapproval.ID
}

func (g fakeApprovalIDGenerator) Next(context.Context) (domainapproval.ID, error) {
	return g.id, nil
}

func TestCreateApproval(t *testing.T) {
	crRepo := &fakeChangeRequestRepository{
		changeRequest: &domainchange.ChangeRequest{
			ID:     "CR00001",
			Status: domainchange.StatusUnderReview,
		},
	}
	approvalRepo := &fakeApprovalRepository{}
	ids := fakeApprovalIDGenerator{id: "APR00001"}

	uc := NewCreateApprovalUseCase(crRepo, approvalRepo, ids)

	got, err := uc.Execute(context.Background(), CreateApprovalCommand{
		ChangeRequestID: "CR00001",
		Type:            domainapproval.TypeSecurity,
		Required:        true,
		ApproverID:      "USR002",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.ID != "APR00001" {
		t.Fatalf("ID = %s, want APR00001", got.ID)
	}
	if got.Status != domainapproval.StatusPending {
		t.Fatalf("Status = %s, want PENDING", got.Status)
	}
	if got.ChangeRequestID != "CR00001" {
		t.Fatalf("ChangeRequestID = %s, want CR00001", got.ChangeRequestID)
	}
	if got.ApproverID != "USR002" {
		t.Fatalf("ApproverID = %s, want USR002", got.ApproverID)
	}
}

func TestCreateApprovalRejectsWrongChangeRequestStatus(t *testing.T) {
	crRepo := &fakeChangeRequestRepository{
		changeRequest: &domainchange.ChangeRequest{
			ID:     "CR00001",
			Status: domainchange.StatusDraft,
		},
	}
	ids := fakeApprovalIDGenerator{id: "APR00001"}

	uc := NewCreateApprovalUseCase(crRepo, &fakeApprovalRepository{}, ids)
	_, err := uc.Execute(context.Background(), CreateApprovalCommand{
		ChangeRequestID: "CR00001",
		Type:            domainapproval.TypeSecurity,
		Required:        true,
		ApproverID:      "USR002",
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want conflict")
	}
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("Execute() error = %v, want ErrConflict", err)
	}
}
