package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	applicationapproval "github.com/balzorn/e5-atlasis/backend/internal/application/approval"
	applicationasset "github.com/balzorn/e5-atlasis/backend/internal/application/asset"
	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

const (
	maxRequestBodySize = 1 << 20
	maxActorIDLength   = 128
	maxChangeCount     = 100
)

type Handler struct {
	getAsset         *applicationasset.GetAssetUseCase
	getChangeRequest *applicationchange.GetChangeRequestUseCase
	createAsset      *applicationasset.CreateAssetUseCase
	createChange     *applicationchange.CreateChangeRequestUseCase
	submitChange     *applicationchange.SubmitChangeRequestUseCase
	startReview      *applicationchange.StartReviewUseCase
	requestChanges   *applicationchange.RequestChangesUseCase
	rejectChange     *applicationchange.RejectChangeRequestUseCase
	approveChange    *applicationchange.ApproveChangeRequestUseCase
	applyChange      *applicationchange.ApplyChangeRequestUseCase
	createApproval   *applicationapproval.CreateApprovalUseCase
	listApprovals    *applicationapproval.ListApprovalsUseCase
	approveApproval  *applicationapproval.ApproveApprovalUseCase
	rejectApproval   *applicationapproval.RejectApprovalUseCase
	principalResolver PrincipalResolver
}

func NewHandler(
	getAsset *applicationasset.GetAssetUseCase,
	getChangeRequest *applicationchange.GetChangeRequestUseCase,
	createAsset *applicationasset.CreateAssetUseCase,
	createChange *applicationchange.CreateChangeRequestUseCase,
	submitChange *applicationchange.SubmitChangeRequestUseCase,
	startReview *applicationchange.StartReviewUseCase,
	requestChanges *applicationchange.RequestChangesUseCase,
	rejectChange *applicationchange.RejectChangeRequestUseCase,
	approveChange *applicationchange.ApproveChangeRequestUseCase,
	applyChange *applicationchange.ApplyChangeRequestUseCase,
	createApproval *applicationapproval.CreateApprovalUseCase,
	listApprovals *applicationapproval.ListApprovalsUseCase,
	approveApproval *applicationapproval.ApproveApprovalUseCase,
	rejectApproval *applicationapproval.RejectApprovalUseCase,
	principalResolver PrincipalResolver,
) *Handler {
	return &Handler{
		getAsset:         getAsset,
		getChangeRequest: getChangeRequest,
		createAsset:      createAsset,
		createChange:     createChange,
		submitChange:     submitChange,
		startReview:      startReview,
		requestChanges:   requestChanges,
		rejectChange:     rejectChange,
		approveChange:    approveChange,
		applyChange:      applyChange,
		createApproval:   createApproval,
		listApprovals:    listApprovals,
		approveApproval:  approveApproval,
		rejectApproval:   rejectApproval,
		principalResolver: principalResolver,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("POST /api/v1/assets", h.createAssetHandler)
	mux.HandleFunc("GET /api/v1/assets/{assetID}", h.getAssetByID)
	mux.HandleFunc("POST /api/v1/change-requests", h.createChangeRequestHandler)
	mux.HandleFunc("GET /api/v1/change-requests/{changeRequestID}", h.getChangeRequestByID)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/submit", h.submitChangeRequestHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/review", h.startChangeRequestReviewHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/request-changes", h.requestChangesHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/reject", h.rejectChangeRequestHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/approve", h.approveChangeRequestHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/apply", h.applyChangeRequestHandler)
	mux.HandleFunc("GET /api/v1/change-requests/{changeRequestID}/approvals", h.listApprovalsHandler)
	mux.HandleFunc("POST /api/v1/change-requests/{changeRequestID}/approvals", h.createApprovalHandler)
	mux.HandleFunc("POST /api/v1/approvals/{approvalID}/approve", h.approveApprovalHandler)
	mux.HandleFunc("POST /api/v1/approvals/{approvalID}/reject", h.rejectApprovalHandler)

	return h.authenticateAPI(mux)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

type createAssetRequest struct {
	Type           domainasset.AssetType       `json:"type"`
	Name           string                      `json:"name"`
	ShortName      string                      `json:"shortName"`
	OrganizationID string                      `json:"organizationId"`
	OwnerID        string                      `json:"ownerId"`
	Purpose        string                      `json:"purpose"`
	Criticality    domainasset.Criticality     `json:"criticality"`
	RiskLevel      domainasset.RiskLevel       `json:"riskLevel"`
	Security       domainasset.SecurityProfile `json:"security"`
}

func (h *Handler) createAssetHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := requireActorID(w, r)
	if !ok {
		return
	}

	var req createAssetRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
			return
		}

		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request")
		return
	}

	created, err := h.createAsset.Execute(r.Context(), applicationasset.CreateAssetCommand{
		Type:           req.Type,
		Name:           req.Name,
		ShortName:      req.ShortName,
		OrganizationID: req.OrganizationID,
		OwnerID:        req.OwnerID,
		Purpose:        req.Purpose,
		Criticality:    req.Criticality,
		RiskLevel:      req.RiskLevel,
		Security:       req.Security,
		CreatedBy:      actorID,
	})
	if err != nil {
		writeApplicationError(w, err)
		return
	}

	location := "/api/v1/assets/" + created.ID.String()
	w.Header().Set("Location", location)
	writeJSON(w, http.StatusCreated, assetResponse(*created))
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return errUnsupportedMediaType
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request must contain exactly one JSON value")
	}

	return nil
}

var errUnsupportedMediaType = errors.New("unsupported media type")

type createChangeRequestValue struct {
	Kind  string          `json:"kind"`
	Value json.RawMessage `json:"value"`
}

type createChangeRequestChange struct {
	Field    string                   `json:"field"`
	NewValue createChangeRequestValue `json:"newValue"`
}

type createChangeRequestRequest struct {
	AssetID     string                      `json:"assetId"`
	BaseVersion int                         `json:"baseVersion"`
	Title       string                      `json:"title"`
	Description string                      `json:"description"`
	Changes     []createChangeRequestChange `json:"changes"`
}

func (h *Handler) createChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	actorID := strings.TrimSpace(r.Header.Get("X-Actor-ID"))
	if actorID == "" || len(actorID) > maxActorIDLength {
		writeError(w, http.StatusBadRequest, "INVALID_ACTOR_ID", "valid X-Actor-ID header is required")
		return
	}

	var req createChangeRequestRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
			return
		}

		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request")
		return
	}

	assetID, err := domainasset.ParseAssetID(strings.TrimSpace(req.AssetID))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ASSET_ID", "invalid asset ID")
		return
	}

	if len(req.Changes) > maxChangeCount {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "too many changes")
		return
	}

	proposals := make([]applicationchange.ChangeProposal, 0, len(req.Changes))
	for _, change := range req.Changes {
		field, err := domainasset.ParseFieldName(strings.TrimSpace(change.Field))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid change field")
			return
		}

		value, err := parseChangeRequestFieldValue(change.NewValue)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid change value")
			return
		}

		proposals = append(proposals, applicationchange.ChangeProposal{
			Field:    field,
			NewValue: value,
		})
	}

	cr, err := h.createChange.Execute(r.Context(), applicationchange.CreateChangeRequestCommand{
		AssetID:     assetID,
		BaseVersion: req.BaseVersion,
		Initiator:   actorID,
		Title:       req.Title,
		Description: req.Description,
		Changes:     proposals,
	})
	if err != nil {
		writeApplicationError(w, err)
		return
	}

	location := "/api/v1/change-requests/" + cr.ID.String()
	w.Header().Set("Location", location)
	writeJSON(w, http.StatusCreated, changeRequestResponse(*cr))
}

func parseChangeRequestFieldValue(value createChangeRequestValue) (domainasset.FieldValue, error) {
	switch value.Kind {
	case string(domainasset.FieldValueKindString):
		var stringValue string
		if len(value.Value) == 0 {
			return domainasset.FieldValue{}, errors.New("string field value is required")
		}
		if err := json.Unmarshal(value.Value, &stringValue); err != nil {
			return domainasset.FieldValue{}, err
		}
		return domainasset.NewStringFieldValue(stringValue), nil

	case string(domainasset.FieldValueKindBoolean):
		var boolValue bool
		if len(value.Value) == 0 {
			return domainasset.FieldValue{}, errors.New("boolean field value is required")
		}
		if err := json.Unmarshal(value.Value, &boolValue); err != nil {
			return domainasset.FieldValue{}, err
		}
		return domainasset.NewBoolFieldValue(boolValue), nil

	default:
		return domainasset.FieldValue{}, errors.New("unknown field value kind")
	}
}

func requireActorID(w http.ResponseWriter, r *http.Request) (string, bool) {
	subject, ok := principalFromContext(r.Context())
	if !ok || strings.TrimSpace(subject.ID) == "" || len(subject.ID) > maxActorIDLength {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "a valid authenticated principal is required")
		return "", false
	}
	return subject.ID, true
}

func parseChangeRequestPathID(w http.ResponseWriter, r *http.Request) (domainchange.ID, bool) {
	id, err := domainchange.ParseChangeRequestID(strings.TrimSpace(r.PathValue("changeRequestID")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CHANGE_REQUEST_ID", "invalid change request ID")
		return "", false
	}
	return id, true
}

func parseApprovalPathID(w http.ResponseWriter, r *http.Request) (domainapproval.ID, bool) {
	id, err := domainapproval.ParseApprovalID(strings.TrimSpace(r.PathValue("approvalID")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_APPROVAL_ID", "invalid approval ID")
		return "", false
	}
	return id, true
}

func writeChangeRequestResult(w http.ResponseWriter, cr *domainchange.ChangeRequest) {
	writeJSON(w, http.StatusOK, changeRequestResponse(*cr))
}

func (h *Handler) submitChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	cr, err := h.submitChange.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeChangeRequestResult(w, cr)
}

func (h *Handler) startChangeRequestReviewHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	cr, err := h.startReview.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeChangeRequestResult(w, cr)
}

func (h *Handler) requestChangesHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	cr, err := h.requestChanges.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeChangeRequestResult(w, cr)
}

func (h *Handler) rejectChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	cr, err := h.rejectChange.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeChangeRequestResult(w, cr)
}

func (h *Handler) approveChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	cr, err := h.approveChange.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeChangeRequestResult(w, cr)
}

func (h *Handler) applyChangeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	id, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	asset, err := h.applyChange.Execute(r.Context(), id)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assetResponse(*asset))
}

type createApprovalRequest struct {
	Type       domainapproval.Type `json:"type"`
	Required   bool                `json:"required"`
	ApproverID string              `json:"approverId"`
}

func (h *Handler) createApprovalHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	crID, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	var req createApprovalRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
			return
		}
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request")
		return
	}
	a, err := h.createApproval.Execute(r.Context(), applicationapproval.CreateApprovalCommand{
		ChangeRequestID: crID, Type: req.Type, Required: req.Required, ApproverID: req.ApproverID,
	})
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	w.Header().Set("Location", "/api/v1/approvals/"+a.ID.String())
	writeJSON(w, http.StatusCreated, approvalResponse(*a))
}

func (h *Handler) listApprovalsHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActorID(w, r); !ok {
		return
	}
	crID, ok := parseChangeRequestPathID(w, r)
	if !ok {
		return
	}
	approvals, err := h.listApprovals.Execute(r.Context(), crID)
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	response := make([]map[string]any, 0, len(approvals))
	for _, a := range approvals {
		response = append(response, approvalResponse(a))
	}
	writeJSON(w, http.StatusOK, response)
}

type approvalDecisionRequest struct {
	Comment string `json:"comment"`
}

func (h *Handler) decideApproval(w http.ResponseWriter, r *http.Request, approve bool) {
	actorID, ok := requireActorID(w, r)
	if !ok {
		return
	}
	approvalID, ok := parseApprovalPathID(w, r)
	if !ok {
		return
	}
	var req approvalDecisionRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json")
			return
		}
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request")
		return
	}
	cmd := applicationapproval.DecisionCommand{ApprovalID: approvalID, DecidedBy: actorID, Comment: strings.TrimSpace(req.Comment)}
	var a *domainapproval.Approval
	var err error
	if approve {
		a, err = h.approveApproval.Execute(r.Context(), cmd)
	} else {
		a, err = h.rejectApproval.Execute(r.Context(), cmd)
	}
	if err != nil {
		writeApplicationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approvalResponse(*a))
}

func (h *Handler) approveApprovalHandler(w http.ResponseWriter, r *http.Request) {
	h.decideApproval(w, r, true)
}
func (h *Handler) rejectApprovalHandler(w http.ResponseWriter, r *http.Request) {
	h.decideApproval(w, r, false)
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
	switch {
	case errors.Is(err, ports.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "operation is not permitted")
	case errors.Is(err, ports.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
	case errors.Is(err, ports.ErrConflict):
		writeError(w, http.StatusConflict, "VERSION_CONFLICT", "resource version conflict")
	case errors.Is(err, ports.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
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

func approvalResponse(a domainapproval.Approval) map[string]any {
	response := map[string]any{
		"id":              a.ID.String(),
		"changeRequestId": a.ChangeRequestID,
		"type":            string(a.Type),
		"status":          string(a.Status),
		"required":        a.Required,
		"approverId":      a.ApproverID,
		"comment":         a.Comment,
	}
	if a.DecidedAt != nil {
		response["decidedAt"] = a.DecidedAt.UTC().Format(time.RFC3339)
	}
	return response
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
