package httpapi

import (
	"bytes"
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
	asset          *domainasset.InformationAsset
	createdAsset   *domainasset.InformationAsset
	createdVersion *domainasset.AssetVersion
	err            error
}

func (r *fakeAssetRepository) Create(
	_ context.Context,
	a domainasset.InformationAsset,
	v domainasset.AssetVersion,
) error {
	r.createdAsset = &a
	r.createdVersion = &v
	return nil
}

func (r *fakeAssetRepository) GetByID(
	_ context.Context,
	_ domainasset.AssetID,
) (*domainasset.InformationAsset, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.asset != nil {
		return r.asset, nil
	}
	if r.createdAsset != nil {
		return r.createdAsset, nil
	}
	return nil, ports.ErrNotFound
}

type fakeAssetIDGenerator struct{}

func (fakeAssetIDGenerator) Next(context.Context) (domainasset.AssetID, error) {
	return domainasset.ParseAssetID("IA00001")
}

type fakeChangeRequestIDGenerator struct{}

func (fakeChangeRequestIDGenerator) Next(context.Context) (domainchange.ID, error) {
	return "CR00001", nil
}

type fakeChangeRequestRepository struct {
	changeRequest *domainchange.ChangeRequest
	created       *domainchange.ChangeRequest
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
	repo := &fakeAssetRepository{
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
	}

	return NewHandler(
		applicationasset.NewGetAssetUseCase(repo),
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
		applicationasset.NewCreateAssetUseCase(repo, fakeAssetIDGenerator{}),
		applicationchange.NewCreateChangeRequestUseCase(
			repo,
			&fakeChangeRequestRepository{},
			fakeChangeRequestIDGenerator{},
		),
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

func TestCreateAsset(t *testing.T) {
	body := strings.NewReader("{\n" +
		"  \"type\": \"information_system\",\n" +
		"  \"name\": \"Registry\",\n" +
		"  \"organizationId\": \"ORG001\",\n" +
		"  \"ownerId\": \"USR001\",\n" +
		"  \"purpose\": \"Information asset registry\",\n" +
		"  \"criticality\": \"HIGH\",\n" +
		"  \"riskLevel\": \"MEDIUM\",\n" +
		"  \"security\": {\n" +
		"    \"protectionRequired\": true,\n" +
		"    \"protectionStatus\": \"IMPLEMENTED\",\n" +
		"    \"attestationStatus\": \"ATTESTED\",\n" +
		"    \"cyberCenterRequired\": true\n" +
		"  }\n" +
		"}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Location"); got != "/api/v1/assets/IA00001" {
		t.Fatalf("Location = %q, want /api/v1/assets/IA00001", got)
	}

	if !strings.Contains(rec.Body.String(), "\"status\":\"DRAFT\"") {
		t.Fatalf("body = %s, want DRAFT status", rec.Body.String())
	}

	if strings.Contains(rec.Body.String(), "\"status\":\"ACTIVE\"") {
		t.Fatalf("body = %s, status must not be client-controlled", rec.Body.String())
	}
}

func TestCreateAssetRejectsUnknownFields(t *testing.T) {
	body := strings.NewReader("{\"type\":\"information_system\",\"name\":\"Registry\",\"organizationId\":\"ORG001\",\"ownerId\":\"USR001\",\"criticality\":\"HIGH\",\"riskLevel\":\"MEDIUM\",\"status\":\"ACTIVE\",\"security\":{\"protectionRequired\":false,\"protectionStatus\":\"NOT_REQUIRED\",\"attestationStatus\":\"NOT_REQUIRED\",\"cyberCenterRequired\":false}}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCreateAssetRejectsMissingActor(t *testing.T) {
	body := bytes.NewBufferString("{\"type\":\"information_system\"}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", body)
	req.Header.Set("Content-Type", "application/json")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCreateAssetRejectsUnsupportedMediaType(t *testing.T) {
	body := bytes.NewBufferString("{}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", body)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", rec.Code)
	}
}

func TestCreateAssetRejectsTrailingJSON(t *testing.T) {
	body := bytes.NewBufferString("{\"type\":\"information_system\"} {\"type\":\"information_system\"}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCreateChangeRequest(t *testing.T) {
	body := strings.NewReader("{\n" +
		"  \"assetId\": \"IA00001\",\n" +
		"  \"baseVersion\": 1,\n" +
		"  \"title\": \"Change owner\",\n" +
		"  \"description\": \"Change the asset owner\",\n" +
		"  \"changes\": [{\n" +
		"    \"field\": \"owner_id\",\n" +
		"    \"newValue\": {\n" +
		"      \"kind\": \"string\",\n" +
		"      \"value\": \"USR002\"\n" +
		"    }\n" +
		"  }]\n" +
		"}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/change-requests", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Location"); got != "/api/v1/change-requests/CR00001" {
		t.Fatalf("Location = %q, want /api/v1/change-requests/CR00001", got)
	}

	bodyText := rec.Body.String()
	for _, fragment := range []string{
		"CR00001",
		"\"assetId\":\"IA00001\"",
		"\"baseVersion\":1",
		"\"status\":\"DRAFT\"",
		"\"field\":\"owner_id\"",
		"\"oldValue\":{\"kind\":\"string\",\"value\":\"USR001\"}",
		"\"newValue\":{\"kind\":\"string\",\"value\":\"USR002\"}",
	} {
		if !strings.Contains(bodyText, fragment) {
			t.Fatalf("body = %s, missing %s", bodyText, fragment)
		}
	}
}

func TestCreateChangeRequestRejectsVersionConflict(t *testing.T) {
	body := strings.NewReader("{\"assetId\":\"IA00001\",\"baseVersion\":2,\"title\":\"Change owner\",\"changes\":[{\"field\":\"owner_id\",\"newValue\":{\"kind\":\"string\",\"value\":\"USR002\"}}]}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/change-requests", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateChangeRequestRejectsInvalidFieldValue(t *testing.T) {
	body := strings.NewReader("{\"assetId\":\"IA00001\",\"baseVersion\":1,\"title\":\"Change owner\",\"changes\":[{\"field\":\"owner_id\",\"newValue\":{\"kind\":\"boolean\",\"value\":true}}]}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/change-requests", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
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
	repo := &fakeAssetRepository{}
	handler := NewHandler(
		applicationasset.NewGetAssetUseCase(repo),
		applicationchange.NewGetChangeRequestUseCase(&fakeChangeRequestRepository{}),
		applicationasset.NewCreateAssetUseCase(repo, fakeAssetIDGenerator{}),
		applicationchange.NewCreateChangeRequestUseCase(
			repo,
			&fakeChangeRequestRepository{},
			fakeChangeRequestIDGenerator{},
		),
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
