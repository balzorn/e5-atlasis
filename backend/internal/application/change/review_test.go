package change

import (
	"context"
	"testing"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

func newReviewTestRepository(
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

func TestStartReview(t *testing.T) {
	repo := newReviewTestRepository(domainchange.StatusSubmitted)
	uc := NewStartReviewUseCase(repo)

	cr, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.Status != domainchange.StatusUnderReview {
		t.Fatalf("Status = %s, want UNDER_REVIEW", cr.Status)
	}
}

func TestRequestChanges(t *testing.T) {
	repo := newReviewTestRepository(domainchange.StatusUnderReview)
	uc := NewRequestChangesUseCase(repo)

	cr, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.Status != domainchange.StatusChangesRequested {
		t.Fatalf("Status = %s, want CHANGES_REQUESTED", cr.Status)
	}
}

func TestRejectChangeRequest(t *testing.T) {
	repo := newReviewTestRepository(domainchange.StatusUnderReview)
	uc := NewRejectChangeRequestUseCase(repo)

	cr, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.Status != domainchange.StatusRejected {
		t.Fatalf("Status = %s, want REJECTED", cr.Status)
	}
}

func TestStartReviewRejectsInvalidStatus(t *testing.T) {
	repo := newReviewTestRepository(domainchange.StatusDraft)
	uc := NewStartReviewUseCase(repo)

	_, err := uc.Execute(context.Background(), "CR00001")
	if err == nil {
		t.Fatal("Execute() error = nil, want transition error")
	}
}
