package change

import (
	"context"
	"testing"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

type fakeAssetRepository struct {
	asset domainasset.InformationAsset
}

func (r fakeAssetRepository) Create(
	context.Context,
	domainasset.InformationAsset,
	domainasset.AssetVersion,
) error {
	return nil
}

func (r fakeAssetRepository) GetByID(
	context.Context,
	domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	return &r.asset, nil
}

type fakeChangeRequestRepository struct {
	created *domainchange.ChangeRequest
}

func (r *fakeChangeRequestRepository) Create(
	_ context.Context,
	cr domainchange.ChangeRequest,
) error {
	r.created = &cr
	return nil
}

func (r *fakeChangeRequestRepository) GetByID(
	context.Context,
	domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return r.created, nil
}

func (r *fakeChangeRequestRepository) Save(
	_ context.Context,
	cr domainchange.ChangeRequest,
) error {
	r.created = &cr
	return nil
}

type fakeChangeRequestIDGenerator struct{}

func (fakeChangeRequestIDGenerator) Next(context.Context) (domainchange.ID, error) {
	return "CR00001", nil
}

func TestCreateChangeRequest(t *testing.T) {
	repo := &fakeChangeRequestRepository{}

	assetID, _ := domainasset.ParseAssetID("IA00001")

	uc := NewCreateChangeRequestUseCase(
		fakeAssetRepository{
			asset: domainasset.InformationAsset{
				ID:             assetID,
				Type:           domainasset.AssetTypeInformationSystem,
				Name:           "Test System",
				OrganizationID: "ORG001",
				OwnerID:        "USR001",
				CurrentVersion: 3,
			},
		},
		repo,
		fakeChangeRequestIDGenerator{},
	)

	cr, err := uc.Execute(context.Background(), CreateChangeRequestCommand{
		AssetID:     assetID,
		BaseVersion: 3,
		Initiator:   "USR002",
		Title:       "Change owner",
		Changes: []ChangeProposal{
			{
				Field:    "owner_id",
				NewValue: "USR003",
			},
		},
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if cr.ID != "CR00001" {
		t.Fatalf("ID = %s, want CR00001", cr.ID)
	}

	if cr.Status != domainchange.StatusDraft {
		t.Fatalf("Status = %s, want DRAFT", cr.Status)
	}

	if got := cr.Changes[0].OldValue; got != "USR001" {
		t.Fatalf("OldValue = %v, want USR001", got)
	}

	if got := cr.Changes[0].NewValue; got != "USR003" {
		t.Fatalf("NewValue = %v, want USR003", got)
	}
}
