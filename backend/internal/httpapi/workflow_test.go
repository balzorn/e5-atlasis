package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateApproval(t *testing.T) {
	body := strings.NewReader("{\n" +
		"  \"type\": \"SECURITY\",\n" +
		"  \"required\": true,\n" +
		"  \"approverId\": \"USR002\"\n" +
		"}")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/change-requests/CR00001/approvals",
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Location"); got != "/api/v1/approvals/APR00003" {
		t.Fatalf("Location = %q, want /api/v1/approvals/APR00003", got)
	}

	for _, fragment := range []string{
		"\"id\":\"APR00003\"",
		"\"changeRequestId\":\"CR00001\"",
		"\"status\":\"PENDING\"",
		"\"approverId\":\"USR002\"",
	} {
		if !strings.Contains(rec.Body.String(), fragment) {
			t.Fatalf("body = %s, missing %s", rec.Body.String(), fragment)
		}
	}
}

func TestListApprovals(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/change-requests/CR00001/approvals",
		nil,
	)
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "\"id\":\"APR00001\"") {
		t.Fatalf("body = %s, missing approval", rec.Body.String())
	}
}

func TestApproveChangeRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/change-requests/CR00001/approve",
		nil,
	)
	req.Header.Set("X-Actor-ID", "USR001")

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "\"status\":\"APPROVED\"") {
		t.Fatalf("body = %s, want APPROVED status", rec.Body.String())
	}
}

func TestChangeRequestCommandRejectsMissingActor(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/change-requests/CR00001/submit",
		nil,
	)

	newTestHandler().Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}
