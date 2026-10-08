package postgres

import (
	"context"
	"testing"
	"time"

	applicationapproval "github.com/balzorn/e5-atlasis/backend/internal/application/approval"
	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

func TestPostgreSQLConcurrentApprovalDecision(t *testing.T) {
	db := newIntegrationDB(t)
	truncateIntegrationTables(t, db)

	assets := NewAssetRepository(db)
	changes := NewChangeRequestRepository(db)
	approvals := NewApprovalRepository(db)

	assetID, err := domainasset.ParseAssetID("IA00004")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	current := domainasset.InformationAsset{
		ID:             assetID,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Approval Concurrency Test System",
		Status:         domainasset.AssetStatusDraft,
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Criticality:    domainasset.CriticalityMedium,
		RiskLevel:      domainasset.RiskLevelMedium,
		Security: domainasset.SecurityProfile{
			ProtectionRequired: false,
			ProtectionStatus:   domainasset.ProtectionStatusNotRequired,
			AttestationStatus:  domainasset.AttestationStatusNotRequired,
		},
		CurrentVersion: 1,
	}

	version := domainasset.AssetVersion{
		ID:        "IA00004-v00001",
		AssetID:   assetID,
		Version:   1,
		State:     current,
		CreatedBy: "USR001",
		CreatedAt: now,
	}

	if err := assets.Create(context.Background(), current, version); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	cr := domainchange.ChangeRequest{
		ID:          "CR00005",
		AssetID:     assetID.String(),
		BaseVersion: 1,
		Status:      domainchange.StatusUnderReview,
		Initiator:   "USR001",
		Title:       "Approval concurrency test",
		Changes: []domainchange.FieldChange{
			{
				ID:       "CHG00001",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR002"),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := changes.Create(context.Background(), cr); err != nil {
		t.Fatalf("create change request: %v", err)
	}

	approval := domainapproval.Approval{
		ID:              "APR00002",
		ChangeRequestID: cr.ID.String(),
		Type:            domainapproval.TypeSecurity,
		Status:          domainapproval.StatusPending,
		Required:        true,
		ApproverID:      "USR002",
	}

	if err := approvals.Create(context.Background(), approval); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	type result struct {
		comment string
		err     error
	}

	results := make(chan result, 2)

	runDecision := func(comment string) {
		uc := applicationapproval.NewApproveApprovalUseCase(approvals)
		_, err := uc.Execute(
			context.Background(),
			applicationapproval.DecisionCommand{
				ApprovalID: approval.ID,
				DecidedBy:  approval.ApproverID,
				Comment:    comment,
			},
		)
		results <- result{comment: comment, err: err}
	}

	go runDecision("decision-1")
	go runDecision("decision-2")

	var successes int
	for i := 0; i < 2; i++ {
		got := <-results
		if got.err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf("successful decisions = %d, want 1", successes)
	}

	stored, err := approvals.GetByID(context.Background(), approval.ID)
	if err != nil {
		t.Fatalf("reload approval: %v", err)
	}

	if stored.Status != domainapproval.StatusApproved {
		t.Fatalf("status = %s, want APPROVED", stored.Status)
	}

	if stored.DecidedAt == nil {
		t.Fatal("decided_at is nil")
	}

	if stored.Comment != "decision-1" && stored.Comment != "decision-2" {
		t.Fatalf("unexpected comment %q", stored.Comment)
	}
}
