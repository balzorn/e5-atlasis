package change

import (
	"context"
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

type fakeChangeApplier struct {
	cr      *domainchange.ChangeRequest
	asset   *domainasset.InformationAsset
	version *domainasset.AssetVersion
}

func (a *fakeChangeApplier) Apply(
	_ context.Context,
	cr domainchange.ChangeRequest,
	asset domainasset.InformationAsset,
	version domainasset.AssetVersion,
) error {
	a.cr = &cr
	a.asset = &asset
	a.version = &version
	return nil
}

func TestApplyChangeRequest(t *testing.T) {
	assetID, _ := domainasset.ParseAssetID("IA00001")

	assets := &fakeAssetRepository{
		asset: domainasset.InformationAsset{
			ID:             assetID,
			Type:           domainasset.AssetTypeInformationSystem,
			Name:           "Test System",
			OrganizationID: "ORG001",
			OwnerID:        "USR001",
			CurrentVersion: 3,
		},
	}

	changes := &fakeChangeRequestRepository{
		created: &domainchange.ChangeRequest{
			ID:          "CR00001",
			AssetID:     "IA00001",
			BaseVersion: 3,
			Status:      domainchange.StatusApproved,
			Initiator:   "USR001",
			Title:       "Change owner",
			Changes: []domainchange.FieldChange{
				{
					ID:       "CHG00001",
					Field:    domainasset.FieldOwnerID,
					OldValue: domainasset.NewStringFieldValue("USR001"),
					NewValue: domainasset.NewStringFieldValue("USR002"),
				},
			},
		},
	}

	applier := &fakeChangeApplier{}

	uc := NewApplyChangeRequestUseCase(
		assets,
		changes,
		applier,
	)

	got, err := uc.Execute(context.Background(), "CR00001")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.OwnerID != "USR002" {
		t.Fatalf("OwnerID = %s, want USR002", got.OwnerID)
	}

	if got.CurrentVersion.Int() != 4 {
		t.Fatalf(
			"CurrentVersion = %d, want 4",
			got.CurrentVersion.Int(),
		)
	}

	if applier.version == nil {
		t.Fatal("asset version was not applied")
	}

	if applier.version.Version.Int() != 4 {
		t.Fatalf(
			"version = %d, want 4",
			applier.version.Version.Int(),
		)
	}

	if applier.cr.Status != domainchange.StatusApplied {
		t.Fatalf(
			"CR status = %s, want APPLIED",
			applier.cr.Status,
		)
	}
}

func TestApplyChangeRequestRejectsStaleVersion(t *testing.T) {
	assetID, _ := domainasset.ParseAssetID("IA00001")

	assets := &fakeAssetRepository{
		asset: domainasset.InformationAsset{
			ID:             assetID,
			Type:           domainasset.AssetTypeInformationSystem,
			Name:           "Test System",
			OrganizationID: "ORG001",
			OwnerID:        "USR001",
			CurrentVersion: 4,
		},
	}

	changes := &fakeChangeRequestRepository{
		created: &domainchange.ChangeRequest{
			ID:          "CR00001",
			AssetID:     "IA00001",
			BaseVersion: 3,
			Status:      domainchange.StatusApproved,
			Initiator:   "USR001",
			Title:       "Stale change",
			Changes: []domainchange.FieldChange{
				{
					ID:       "CHG00001",
					Field:    domainasset.FieldOwnerID,
					OldValue: domainasset.NewStringFieldValue("USR001"),
					NewValue: domainasset.NewStringFieldValue("USR002"),
				},
			},
		},
	}

	uc := NewApplyChangeRequestUseCase(
		assets,
		changes,
		&fakeChangeApplier{},
	)

	_, err := uc.Execute(context.Background(), "CR00001")
	if err == nil {
		t.Fatal("Execute() error = nil, want stale version error")
	}
}
