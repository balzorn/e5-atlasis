package asset

import (
	"context"
	"errors"
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
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

type fakeAssetIDGenerator struct {
	called bool
}

func (g *fakeAssetIDGenerator) Next(context.Context) (domainasset.AssetID, error) {
	g.called = true
	return domainasset.ParseAssetID("IA00001")
}

func TestCreateAsset(t *testing.T) {
	repo := &fakeAssetRepository{}
	ids := &fakeAssetIDGenerator{}
	uc := NewCreateAssetUseCase(repo, ids)

	got, err := uc.Execute(context.Background(), CreateAssetCommand{
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Test System",
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Purpose:        "Testing",
		Criticality:    domainasset.CriticalityHigh,
		RiskLevel:      domainasset.RiskLevelMedium,
		Security: domainasset.SecurityProfile{
			ProtectionRequired:  false,
			ProtectionStatus:   domainasset.ProtectionStatusNotRequired,
			AttestationStatus:  domainasset.AttestationStatusNotRequired,
			CyberCenterRequired: false,
		},
		CreatedBy: "USR001",
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

	if repo.createdVersion.CreatedBy != "USR001" {
		t.Fatalf("created by = %q, want USR001", repo.createdVersion.CreatedBy)
	}

	if repo.createdVersion.CreatedAt.IsZero() {
		t.Fatal("created at must be set")
	}
}

func TestCreateAssetRejectsInvalidCommandBeforeIDAllocation(t *testing.T) {
	ids := &fakeAssetIDGenerator{}
	uc := NewCreateAssetUseCase(&fakeAssetRepository{}, ids)

	_, err := uc.Execute(context.Background(), CreateAssetCommand{
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           " ",
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Criticality:    domainasset.CriticalityMedium,
		RiskLevel:      domainasset.RiskLevelMedium,
		Security: domainasset.SecurityProfile{
			ProtectionStatus:  domainasset.ProtectionStatusNotRequired,
			AttestationStatus: domainasset.AttestationStatusNotRequired,
		},
		CreatedBy: "USR001",
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}

	if !errors.Is(err, ports.ErrInvalidInput) {
		t.Fatalf("Execute() error = %v, want ErrInvalidInput", err)
	}

	if ids.called {
		t.Fatal("ID generator was called for invalid command")
	}
}
