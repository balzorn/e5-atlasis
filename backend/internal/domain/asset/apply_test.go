package asset

import "testing"

func newTestAsset() InformationAsset {
	id, _ := ParseAssetID("IA00001")

	return InformationAsset{
		ID:             id,
		Type:           AssetTypeInformationSystem,
		Name:           "Test System",
		ShortName:      "TEST",
		Status:         AssetStatusDraft,
		OrganizationID: "ORG001",
		OwnerID:        "USR001",
		Purpose:        "Testing",
		Criticality:    CriticalityMedium,
		RiskLevel:      RiskLevelLow,
	}
}

func TestApplyFieldOwner(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldOwnerID,
		NewStringFieldValue("USR002"),
	)
	if err != nil {
		t.Fatalf("ApplyField() error = %v", err)
	}

	if a.OwnerID != "USR002" {
		t.Fatalf("OwnerID = %s, want USR002", a.OwnerID)
	}
}

func TestApplyFieldCriticality(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldCriticality,
		NewStringFieldValue("HIGH"),
	)
	if err != nil {
		t.Fatalf("ApplyField() error = %v", err)
	}

	if a.Criticality != CriticalityHigh {
		t.Fatalf("Criticality = %s, want HIGH", a.Criticality)
	}
}

func TestApplyFieldBoolean(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldProtectionRequired,
		NewBoolFieldValue(true),
	)
	if err != nil {
		t.Fatalf("ApplyField() error = %v", err)
	}

	if !a.Security.ProtectionRequired {
		t.Fatal("ProtectionRequired = false, want true")
	}
}

func TestApplyFieldStatus(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldStatus,
		NewStringFieldValue("ACTIVE"),
	)
	if err != nil {
		t.Fatalf("ApplyField() error = %v", err)
	}

	if a.Status != AssetStatusActive {
		t.Fatalf("Status = %s, want ACTIVE", a.Status)
	}
}

func TestApplyFieldRejectsInvalidStatusTransition(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldStatus,
		NewStringFieldValue("SUSPENDED"),
	)
	if err == nil {
		t.Fatal("ApplyField() error = nil, want transition error")
	}
}

func TestApplyFieldRejectsEmptyName(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldNameField,
		NewStringFieldValue(""),
	)
	if err == nil {
		t.Fatal("ApplyField() error = nil, want required-field error")
	}
}

func TestApplyFieldRejectsEmptyOwner(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldOwnerID,
		NewStringFieldValue(""),
	)
	if err == nil {
		t.Fatal("ApplyField() error = nil, want required-field error")
	}
}

func TestApplyFieldRejectsWrongValueKind(t *testing.T) {
	a := newTestAsset()

	err := a.ApplyField(
		FieldProtectionRequired,
		NewStringFieldValue("true"),
	)
	if err == nil {
		t.Fatal("ApplyField() error = nil, want value kind error")
	}
}
