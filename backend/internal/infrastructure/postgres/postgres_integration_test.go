package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	applicationapproval "github.com/balzorn/e5-atlasis/backend/internal/application/approval"
	applicationchange "github.com/balzorn/e5-atlasis/backend/internal/application/change"
	domainapproval "github.com/balzorn/e5-atlasis/backend/internal/domain/approval"
	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
)

const integrationDatabaseURL = "postgres://e5_atlasis:e5_atlasis_dev@127.0.0.1:54329/e5_atlasis?sslmode=disable"

func TestPostgreSQLFullChangeApplication(t *testing.T) {

	db := newIntegrationDB(t)
	truncateIntegrationTables(t, db)

	assets := NewAssetRepository(db)
	changes := NewChangeRequestRepository(db)
	approvals := NewApprovalRepository(db)
	applier := NewChangeApplier(db)

	assetID, err := domainasset.ParseAssetID("IA00001")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	current := domainasset.InformationAsset{
		ID:             assetID,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Integration Test System",
		ShortName:      "IT-TEST",
		Status:         domainasset.AssetStatusDraft,
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Purpose:        "PostgreSQL integration test",
		Criticality:    domainasset.CriticalityMedium,
		RiskLevel:      domainasset.RiskLevelMedium,
		Security: domainasset.SecurityProfile{
			ProtectionRequired:  true,
			ProtectionStatus:    domainasset.ProtectionStatusRequired,
			AttestationStatus:   domainasset.AttestationStatusRequired,
			CyberCenterRequired: true,
		},
		CurrentVersion: 1,
	}

	if err := current.Validate(); err != nil {
		t.Fatal(err)
	}

	version1 := domainasset.AssetVersion{
		ID:        "IA00001-v00001",
		AssetID:   assetID,
		Version:   1,
		State:     current,
		CreatedBy: "USR001",
		CreatedAt: now,
	}

	if err := assets.Create(context.Background(), current, version1); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	cr := domainchange.ChangeRequest{
		ID:          "CR00001",
		AssetID:     assetID.String(),
		BaseVersion: 1,
		Status:      domainchange.StatusUnderReview,
		Initiator:   "USR001",
		Title:       "Change owner",
		Description: "Integration test change request",
		Changes: []domainchange.FieldChange{
			{
				ID:       "CHG00001",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR002"),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := cr.Validate(); err != nil {
		t.Fatal(err)
	}

	if err := changes.Create(context.Background(), cr); err != nil {
		t.Fatalf("create change request: %v", err)
	}

	approval := domainapproval.Approval{
		ID:              "APR00001",
		ChangeRequestID: cr.ID.String(),
		Type:            domainapproval.TypeSecurity,
		Status:          domainapproval.StatusPending,
		Required:        true,
		ApproverID:      "USR002",
	}

	if err := approvals.Create(context.Background(), approval); err != nil {
		t.Fatalf("create approval: %v", err)
	}

	approveApprovalUC := applicationapproval.NewApproveApprovalUseCase(approvals)
	if _, err := approveApprovalUC.Execute(context.Background(), applicationapproval.DecisionCommand{
		ApprovalID: "APR00001",
		DecidedBy:  "USR002",
		Comment:    "Approved in integration test",
	}); err != nil {
		t.Fatalf("approve approval: %v", err)
	}

	approveCRUC := applicationchange.NewApproveChangeRequestUseCase(changes, approvals)
	approvedCR, err := approveCRUC.Execute(context.Background(), cr.ID)
	if err != nil {
		t.Fatalf("approve change request: %v", err)
	}

	if approvedCR.Status != domainchange.StatusApproved {
		t.Fatalf("CR status = %s, want APPROVED", approvedCR.Status)
	}

	applyUC := applicationchange.NewApplyChangeRequestUseCase(
		assets,
		changes,
		applier,
	)

	got, err := applyUC.Execute(context.Background(), cr.ID)
	if err != nil {
		t.Fatalf("apply change request: %v", err)
	}

	if got.OwnerID != "USR002" {
		t.Fatalf("owner = %s, want USR002", got.OwnerID)
	}

	if got.CurrentVersion.Int() != 2 {
		t.Fatalf("current version = %d, want 2", got.CurrentVersion.Int())
	}

	storedAsset, err := assets.GetByID(context.Background(), assetID)
	if err != nil {
		t.Fatalf("reload asset: %v", err)
	}

	if storedAsset.CurrentVersion.Int() != 2 {
		t.Fatalf("stored current version = %d, want 2", storedAsset.CurrentVersion.Int())
	}

	if storedAsset.OwnerID != "USR002" {
		t.Fatalf("stored owner = %s, want USR002", storedAsset.OwnerID)
	}

	storedCR, err := changes.GetByID(context.Background(), cr.ID)
	if err != nil {
		t.Fatalf("reload change request: %v", err)
	}

	if storedCR.Status != domainchange.StatusApplied {
		t.Fatalf("stored CR status = %s, want APPLIED", storedCR.Status)
	}

	storedApproval, err := approvals.GetByID(context.Background(), approval.ID)
	if err != nil {
		t.Fatalf("reload approval: %v", err)
	}

	if storedApproval.Status != domainapproval.StatusApproved {
		t.Fatalf("stored approval status = %s, want APPROVED", storedApproval.Status)
	}

	var versionCount int
	if err := db.pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM information_asset_versions WHERE asset_id = $1",
		assetID.String(),
	).Scan(&versionCount); err != nil {
		t.Fatalf("count asset versions: %v", err)
	}

	if versionCount != 2 {
		t.Fatalf("asset version count = %d, want 2", versionCount)
	}

	var (
		version2Owner string
		changeRequestID *string
	)

	if err := db.pool.QueryRow(
		context.Background(),
		`SELECT owner_id, change_request_id
		   FROM information_asset_versions
		  WHERE asset_id = $1 AND version = 2`,
		assetID.String(),
	).Scan(&version2Owner, &changeRequestID); err != nil {
		t.Fatalf("read version 2: %v", err)
	}

	if version2Owner != "USR002" {
		t.Fatalf("version 2 owner = %s, want USR002", version2Owner)
	}

	if changeRequestID == nil || *changeRequestID != cr.ID.String() {
		t.Fatalf("version 2 change request = %v, want %s", changeRequestID, cr.ID)
	}
}

func TestPostgreSQLChangeApplierRollsBackOnCRStateMismatch(t *testing.T) {

	db := newIntegrationDB(t)
	truncateIntegrationTables(t, db)

	assets := NewAssetRepository(db)
	applier := NewChangeApplier(db)

	assetID, err := domainasset.ParseAssetID("IA00002")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	current := domainasset.InformationAsset{
		ID:             assetID,
		Type:           domainasset.AssetTypeInformationSystem,
		Name:           "Rollback Test System",
		Status:         domainasset.AssetStatusDraft,
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Criticality:    domainasset.CriticalityLow,
		RiskLevel:      domainasset.RiskLevelLow,
		Security: domainasset.SecurityProfile{
			ProtectionRequired:  false,
			ProtectionStatus:   domainasset.ProtectionStatusNotRequired,
			AttestationStatus:  domainasset.AttestationStatusNotRequired,
		},
		CurrentVersion: 1,
	}

	version1 := domainasset.AssetVersion{
		ID:        "IA00002-v00001",
		AssetID:   assetID,
		Version:   1,
		State:     current,
		CreatedBy: "USR001",
		CreatedAt: now,
	}

	if err := assets.Create(context.Background(), current, version1); err != nil {
		t.Fatalf("create asset: %v", err)
	}

	cr := domainchange.ChangeRequest{
		ID:          "CR00002",
		AssetID:     assetID.String(),
		BaseVersion: 1,
		Status:      domainchange.StatusApplied,
		Initiator:   "USR001",
		Title:       "Invalid direct application",
		Changes: []domainchange.FieldChange{
			{
				ID:       "CHG00001",
				Field:    domainasset.FieldOwnerID,
				OldValue: domainasset.NewStringFieldValue("USR001"),
				NewValue: domainasset.NewStringFieldValue("USR002"),
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	stubVersion2 := current
	stubVersion2.OwnerID = "USR002"
	stubVersion2.CurrentVersion = 2

	version2 := domainasset.AssetVersion{
		ID:              "IA00002-v00002",
		AssetID:         assetID,
		Version:         2,
		State:           stubVersion2,
		CreatedBy:       "USR001",
		CreatedAt:       now,
		ChangeRequestID: stringPointer(cr.ID.String()),
	}

	// Insert a matching CR in DRAFT state so the final status guard fails.
	cr.Status = domainchange.StatusDraft
	if err := NewChangeRequestRepository(db).Create(context.Background(), cr); err != nil {
		t.Fatalf("create change request: %v", err)
	}
	cr.Status = domainchange.StatusApplied

	if err := applier.Apply(context.Background(), cr, stubVersion2, version2); err == nil {
		t.Fatal("Apply() error = nil, want state mismatch error")
	}

	var currentVersion int
	if err := db.pool.QueryRow(
		context.Background(),
		"SELECT current_version FROM information_assets WHERE id = $1",
		assetID.String(),
	).Scan(&currentVersion); err != nil {
		t.Fatalf("read current version: %v", err)
	}

	if currentVersion != 1 {
		t.Fatalf("current version after rollback = %d, want 1", currentVersion)
	}

	var versionCount int
	if err := db.pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM information_asset_versions WHERE asset_id = $1",
		assetID.String(),
	).Scan(&versionCount); err != nil {
		t.Fatalf("count versions after rollback: %v", err)
	}

	if versionCount != 1 {
		t.Fatalf("version count after rollback = %d, want 1", versionCount)
	}
}

func newIntegrationDB(t *testing.T) *DB {
	t.Helper()

	databaseURL := os.Getenv("E5_ATLASIS_TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = integrationDatabaseURL
	}

	db, err := New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	t.Cleanup(db.Close)

	return db
}

func truncateIntegrationTables(t *testing.T, db *DB) {
	t.Helper()

	_, err := db.pool.Exec(
		context.Background(),
		`TRUNCATE TABLE
			approvals,
			change_request_changes,
			information_asset_versions,
			change_requests,
			information_assets
		 CASCADE`,
	)
	if err != nil {
		t.Fatalf("truncate integration tables: %v", err)
	}
}
