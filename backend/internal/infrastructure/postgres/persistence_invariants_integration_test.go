package postgres

import (
	"context"
	"testing"
	"time"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

func TestPostgreSQLRejectsDuplicateFieldChanges(t *testing.T) {
	db := newIntegrationDB(t)
	truncateIntegrationTables(t, db)

	assets := NewAssetRepository(db)
	changes := NewChangeRequestRepository(db)

	assetID, err := domainasset.ParseAssetID("IA00005")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	current := domainasset.InformationAsset{
		ID:             assetID,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Duplicate Field Test System",
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
		ID:        "IA00005-v00001",
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
		ID:          "CR00006",
		AssetID:     assetID.String(),
		BaseVersion: 1,
		Status:      domainchange.StatusDraft,
		Initiator:   "USR001",
		Title:       "Duplicate field test",
		Changes: []domainchange.FieldChange{
			{
				ID:       "CHG00001",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR002"),
			},
			{
				ID:       "CHG00002",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR003"),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := changes.Create(context.Background(), cr); err == nil {
		t.Fatal("Create() error = nil, want duplicate field constraint error")
	}

	var count int
	if err := db.pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM change_requests WHERE id = $1",
		cr.ID.String(),
	).Scan(&count); err != nil {
		t.Fatalf("count change request: %v", err)
	}

	if count != 0 {
		t.Fatalf("change request count = %d, want 0", count)
	}
}
