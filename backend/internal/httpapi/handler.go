package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	applicationasset "github.com/balzorn/e5-atlasis/backend/internal/application/asset"
	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type Handler struct {
	getAsset         *applicationasset.GetAssetUseCase
	getChangeRequest *applicationchange.GetChangeRequestUseCase
}

func NewHandler(
	getAsset *applicationasset.GetAssetUseCase,
	getChangeRequest *applicationchange.GetChangeRequestUseCase,
) *Handler {
	return &Handler{
		getAsset:         getAsset,
		getChangeRequest: getChangeRequest,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /api/v1/assets/{assetID}", h.getAssetByID)
	mux.HandleFunc("GET /api/v1/change-requests/{changeRequestID}", h.getChangeRequestByID)

	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *Handler) getAssetByID(w http.ResponseWriter, r *http.Request) {
	value := strings.TrimSpace(r.PathValue("assetID"))
	assetID, err := domainasset.ParseAssetID(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ASSET_ID", "invalid asset ID")
		return
	}

	a, err := h.getAsset.Execute(r.Context(), assetID)
	if err != nil {
		writeApplicationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, assetResponse(*a))
}

func (h *Handler) getChangeRequestByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("changeRequestID"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_CHANGE_REQUEST_ID", "change request ID is required")
		return
	}

	cr, err := h.getChangeRequest.Execute(r.Context(), domainchange.ID(id))
	if err != nil {
		writeApplicationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, changeRequestResponse(*cr))
}

func writeApplicationError(w http.ResponseWriter, err error) {
	if errors.Is(err, ports.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		return
	}

	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"code":    code,
		"message": message,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func assetResponse(a domainasset.InformationAsset) map[string]any {
	return map[string]any{
		"id":             a.ID.String(),
		"type":           string(a.Type),
		"name":           a.Name,
		"shortName":      a.ShortName,
		"status":         string(a.Status),
		"organizationId": a.OrganizationID,
		"ownerId":        a.OwnerID,
		"purpose":        a.Purpose,
		"criticality":    string(a.Criticality),
		"riskLevel":      string(a.RiskLevel),
		"security": map[string]any{
			"protectionRequired":  a.Security.ProtectionRequired,
			"protectionStatus":    string(a.Security.ProtectionStatus),
			"attestationStatus":   string(a.Security.AttestationStatus),
			"cyberCenterRequired": a.Security.CyberCenterRequired,
		},
		"currentVersion": a.CurrentVersion.Int(),
	}
}

func changeRequestResponse(cr domainchange.ChangeRequest) map[string]any {
	changes := make([]map[string]any, 0, len(cr.Changes))

	for _, c := range cr.Changes {
		changes = append(changes, map[string]any{
			"id":       c.ID,
			"field":    string(c.Field),
			"oldValue": fieldValueResponse(c.OldValue),
			"newValue": fieldValueResponse(c.NewValue),
		})
	}

	return map[string]any{
		"id":          cr.ID.String(),
		"assetId":     cr.AssetID,
		"baseVersion": cr.BaseVersion,
		"status":      string(cr.Status),
		"initiator":   cr.Initiator,
		"title":       cr.Title,
		"description": cr.Description,
		"changes":     changes,
		"createdAt":   cr.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":   cr.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func fieldValueResponse(v domainasset.FieldValue) map[string]any {
	switch v.Kind() {
	case domainasset.FieldValueKindString:
		value, _ := v.StringValue()
		return map[string]any{
			"kind":  string(v.Kind()),
			"value": value,
		}
	case domainasset.FieldValueKindBoolean:
		value, _ := v.BoolValue()
		return map[string]any{
			"kind":  string(v.Kind()),
			"value": value,
		}
	default:
		return map[string]any{
			"kind": string(v.Kind()),
		}
	}
}
