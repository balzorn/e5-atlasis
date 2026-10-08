package change

import (
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

func TestChangeRequestTransitions(t *testing.T) {
	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		{StatusDraft, StatusSubmitted, true},
		{StatusSubmitted, StatusUnderReview, true},
		{StatusUnderReview, StatusChangesRequested, true},
		{StatusUnderReview, StatusApproved, true},
		{StatusUnderReview, StatusRejected, true},
		{StatusChangesRequested, StatusUnderReview, true},
		{StatusApproved, StatusApplying, true},
		{StatusApplying, StatusApplied, true},

		{StatusApplied, StatusDraft, false},
		{StatusRejected, StatusApproved, false},
		{StatusCancelled, StatusSubmitted, false},
	}

	var cr ChangeRequest

	for _, tt := range tests {
		cr.Status = tt.from

		if got := cr.CanTransitionTo(tt.to); got != tt.want {
			t.Errorf(
				"CanTransitionTo(%s -> %s) = %v, want %v",
				tt.from,
				tt.to,
				got,
				tt.want,
			)
		}
	}
}

func TestChangeRequestValidation(t *testing.T) {
	cr := ChangeRequest{
		ID:          "CR00001",
		AssetID:     "IA00001",
		BaseVersion: 1,
		Initiator:   "USR00001",
		Title:       "Change owner",
		Changes: []FieldChange{
			{
				ID:       "CHG00001",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR002"),
			},
		},
	}

	if err := cr.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
