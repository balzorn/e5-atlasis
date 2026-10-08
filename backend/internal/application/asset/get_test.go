package asset

import (
	"context"
	"errors"
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type fakeAssetGetRepository struct {
	asset *domainasset.InformationAsset
	err   error
}

func (r *fakeAssetGetRepository) Create(
	_ context.Context,
	_ domainasset.InformationAsset,
	_ domainasset.AssetVersion,
) error {
	return nil
}

func (r *fakeAssetGetRepository) GetByID(
	_ context.Context,
	_ domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	return r.asset, r.err
}

func TestGetAsset(t *testing.T) {
	id, _ := domainasset.ParseAssetID("IA00001")
	expected := &domainasset.InformationAsset{
		ID:             id,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Test System",
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		CurrentVersion: 1,
	}

	uc := NewGetAssetUseCase(&fakeAssetGetRepository{asset: expected})

	got, err := uc.Execute(context.Background(), id)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got != expected {
		t.Fatal("Execute() returned unexpected asset")
	}
}

func TestGetAssetNotFound(t *testing.T) {
	id, _ := domainasset.ParseAssetID("IA00001")
	uc := NewGetAssetUseCase(&fakeAssetGetRepository{})

	_, err := uc.Execute(context.Background(), id)
	if !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("Execute() error = %v, want ErrNotFound", err)
	}
}
