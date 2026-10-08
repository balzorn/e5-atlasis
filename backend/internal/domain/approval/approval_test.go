package approval

import "testing"

func TestApprovalValidation(t *testing.T) {
	a := Approval{
		ID:              "APR00001",
		ChangeRequestID: "CR00001",
		Type:            TypeAssetOwner,
		Status:          StatusPending,
		Required:        true,
		ApproverID:      "USR00001",
	}

	if err := a.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
