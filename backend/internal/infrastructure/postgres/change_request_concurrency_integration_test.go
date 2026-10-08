package postgres

import (
	"context"
	"testing"
	"time"

	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

func TestPostgreSQLConcurrentChangeRequestTransition(t *testing.T) {
	db := newIntegrationDB(t)
	truncateIntegrationTables(t, db)

	assets := NewAssetRepository(db)
	changes := NewChangeRequestRepository(db)

	assetID, err := domainasset.ParseAssetID("IA00001")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	current := domainasset.InformationAsset{
		ID:             assetID,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "CR Transition Concurrency Test System",
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
		ID:        "IA00001-v00001",
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
		ID:          "CR00001",
		AssetID:     assetID.String(),
		BaseVersion: 1,
		Status:      domainchange.StatusDraft,
		Initiator:   "USR001",
		Title:       "Concurrent submit",
		Changes: []domainchange.FieldChange{{
			ID:       "CHG00001",
			Field:    domainasset.FieldOwnerID,
			OldValue: domainasset.NewStringFieldValue("USR001"),
			NewValue: domainasset.NewStringFieldValue("USR002"),
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := changes.Create(context.Background(), cr); err != nil {
		t.Fatalf("create change request: %v", err)
	}

	uc := applicationchange.NewSubmitChangeRequestUseCase(changes)
	results := make(chan error, 2)

	go func() {
		_, err := uc.Execute(context.Background(), cr.ID)
		results <- err
	}()
	go func() {
		_, err := uc.Execute(context.Background(), cr.ID)
		results <- err
	}()

	var successes int
	var conflicts int
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
		} else if isConflictError(err) {
			conflicts++
		} else {
			t.Fatalf("unexpected error = %v", err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes = %d, conflicts = %d, want 1 and 1", successes, conflicts)
	}

	stored, err := changes.GetByID(context.Background(), cr.ID)
	if err != nil {
		t.Fatalf("reload change request: %v", err)
	}
	if stored.Status != domainchange.StatusSubmitted {
		t.Fatalf("stored status = %s, want SUBMITTED", stored.Status)
	}
}

func isConflictError(err error) bool {
	return errors.Is(err, ports.ErrConflict)
}
