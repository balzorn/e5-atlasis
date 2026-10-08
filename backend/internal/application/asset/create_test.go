package asset

import (
	"context"
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

type fakeAssetRepository struct {
	createdAsset   domainasset.InformationAsset
	createdVersion domainasset.AssetVersion
}

func (r *fakeAssetRepository) Create(
	_ context.Context,
	a domainasset.InformationAsset,
	v domainasset.AssetVersion,
) error {
	r.createdAsset = a
	r.createdVersion = v
	return nil
}

func (r *fakeAssetRepository) GetByID(
	_ context.Context,
	_ domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	return nil, nil
}

type fakeAssetIDGenerator struct{}

func (fakeAssetIDGenerator) Next(context.Context) (domainasset.AssetID, error) {
	return domainasset.ParseAssetID("IA00001")
}

func TestCreateAsset(t *testing.T) {
	repo := &fakeAssetRepository{}
	uc := NewCreateAssetUseCase(repo, fakeAssetIDGenerator{})

	got, err := uc.Execute(context.Background(), CreateAssetCommand{
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Test System",
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Purpose:        "Testing",
		Criticality:    domainasset.CriticalityHigh,
		RiskLevel:      domainasset.RiskLevelMedium,
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.ID != "IA00001" {
		t.Fatalf("ID = %s, want IA00001", got.ID)
	}

	if got.Status != domainasset.AssetStatusDraft {
		t.Fatalf("Status = %s, want DRAFT", got.Status)
	}

	if got.CurrentVersion != 1 {
		t.Fatalf("CurrentVersion = %d, want 1", got.CurrentVersion)
	}

	if repo.createdVersion.Version != 1 {
		t.Fatalf("created version = %d, want 1", repo.createdVersion.Version)
	}
}
