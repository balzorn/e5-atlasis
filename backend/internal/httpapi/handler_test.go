package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	applicationasset "github.com/balzorn/e5-atlasis/backend/internal/application/asset"
	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type fakeAssetRepository struct {
	asset *domainasset.InformationAsset
	err   error
}

func (r *fakeAssetRepository) Create(
	_ context.Context,
	_ domainasset.InformationAsset,
	_ domainasset.AssetVersion,
) error {
	return nil
}

func (r *fakeAssetRepository) GetByID(
	_ context.Context,
	_ domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	return r.asset, r.err
}

type fakeChangeRequestRepository struct {
	changeRequest *domainchange.ChangeRequest
	err           error
}

func (r *fakeChangeRequestRepository) Create(
	_ context.Context,
	cr domainchange.ChangeRequest,
) error {
	r.changeRequest = &cr
	return nil
}

func (r *fakeChangeRequestRepository) GetByID(
	_ context.Context,
	_ domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return r.changeRequest, r.err
}

func (r *fakeChangeRequestRepository) Save(
	_ context.Context,
	_ domainchange.ChangeRequest,
) error {
	return nil
}

func newTestHandler() *Handler {
	return NewHandler(
		applicationasset.NewGetAssetUseCase(&fakeAssetRepository{
			asset: &domainasset.InformationAsset{
				ID:             "IA00001",
				Type:           domainasset.AssetTypeInformationSystem,
				Name:           "Test System",
				ShortName:      "TEST",
				Status:         domainasset.AssetStatusDraft,
				OrganizationID: "ORG001",
				OwnerID:        "USR001",
				Purpose:        "Test",
				Criticality:    domainasset.CriticalityMedium,
				RiskLevel:      domainasset.RiskLevelMedium,
				Security: domainasset.SecurityProfile{
					ProtectionRequired:  false,
					ProtectionStatus:   domainasset.ProtectionStatusNotRequired,
					AttestationStatus:  domainasset.AttestationStatusNotRequired,
					CyberCenterRequired: false,
				},
				CurrentVersion: 1,
			},
		}),
		applicationchange.NewGetChangeRequestUseCase(&fakeChangeRequestRepository{
			changeRequest: &domainchange.ChangeRequest{
				ID:          "CR00001",
				AssetID:     "IA00001",
				BaseVersion: 1,
				Status:      domainchange.StatusUnderReview,
				Initiator:   "USR001",
				Title:       "Test change",
			},
		}),
	)
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "status") {
		t.Fatalf("body = %s, want health response", rec.Body.String())
	}
}

func TestGetAssetByID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/IA00001", nil)

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "IA00001") {
		t.Fatalf("body = %s, missing asset ID", body)
	}
	if !strings.Contains(body, "currentVersion") {
		t.Fatalf("body = %s, missing version", body)
	}
}

func TestGetAssetByIDRejectsInvalidID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/not-an-asset", nil)

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetAssetByIDReturnsNotFound(t *testing.T) {
	handler := NewHandler(
		applicationasset.NewGetAssetUseCase(&fakeAssetRepository{}),
		applicationchange.NewGetChangeRequestUseCase(&fakeChangeRequestRepository{}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/IA00099", nil)

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestGetChangeRequestByID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/change-requests/CR00001", nil)

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "CR00001") {
		t.Fatalf("body = %s, missing change request ID", rec.Body.String())
	}
}

var _ ports.AssetRepository = (*fakeAssetRepository)(nil)
var _ ports.ChangeRequestRepository = (*fakeChangeRequestRepository)(nil)
