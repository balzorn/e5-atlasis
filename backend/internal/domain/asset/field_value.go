package asset

import "fmt"

type FieldValueKind string

const (
	FieldValueKindString  FieldValueKind = "string"
	FieldValueKindBoolean FieldValueKind = "boolean"
)

type FieldValue struct {
	kind        FieldValueKind
	stringValue string
	boolValue   bool
}

func NewStringFieldValue(value string) FieldValue {
	return FieldValue{
		kind:        FieldValueKindString,
		stringValue: value,
	}
}

func NewBoolFieldValue(value bool) FieldValue {
	return FieldValue{
		kind:      FieldValueKindBoolean,
		boolValue: value,
	}
}

func (v FieldValue) Kind() FieldValueKind {
	return v.kind
}

func (v FieldValue) StringValue() (string, error) {
	if v.kind != FieldValueKindString {
		return "", fmt.Errorf(
			"field value has kind %q, expected string",
			v.kind,
		)
	}

	return v.stringValue, nil
}

func (v FieldValue) BoolValue() (bool, error) {
	if v.kind != FieldValueKindBoolean {
		return false, fmt.Errorf(
			"field value has kind %q, expected boolean",
			v.kind,
		)
	}

	return v.boolValue, nil
}

func (v FieldValue) Equal(other FieldValue) bool {
	if v.kind != other.kind {
		return false
	}

	switch v.kind {
	case FieldValueKindString:
		return v.stringValue == other.stringValue

	case FieldValueKindBoolean:
		return v.boolValue == other.boolValue

	default:
		return false
	}
}

func (v FieldValue) ValidateFor(field FieldName) error {
	switch field {
	case FieldType:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch AssetType(value) {
		case AssetTypeInformationSystem,
			AssetTypeInformationInfrastructureObject:
			return nil
		default:
			return fmt.Errorf("invalid asset type %q", value)
		}

	case FieldStatus:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch AssetStatus(value) {
		case AssetStatusDraft,
			AssetStatusActive,
			AssetStatusSuspended,
			AssetStatusRetired:
			return nil
		default:
			return fmt.Errorf("invalid asset status %q", value)
		}

	case FieldCriticality:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch Criticality(value) {
		case CriticalityLow,
			CriticalityMedium,
			CriticalityHigh,
			CriticalityCritical:
			return nil
		default:
			return fmt.Errorf("invalid criticality %q", value)
		}

	case FieldRiskLevel:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch RiskLevel(value) {
		case RiskLevelLow,
			RiskLevelMedium,
			RiskLevelHigh,
			RiskLevelCritical:
			return nil
		default:
			return fmt.Errorf("invalid risk level %q", value)
		}

	case FieldProtectionStatus:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch ProtectionStatus(value) {
		case ProtectionStatusNotRequired,
			ProtectionStatusRequired,
			ProtectionStatusInProgress,
			ProtectionStatusImplemented:
			return nil
		default:
			return fmt.Errorf("invalid protection status %q", value)
		}

	case FieldAttestationStatus:
		value, err := v.StringValue()
		if err != nil {
			return err
		}

		switch AttestationStatus(value) {
		case AttestationStatusNotRequired,
			AttestationStatusRequired,
			AttestationStatusInProgress,
			AttestationStatusAttested,
			AttestationStatusExpired:
			return nil
		default:
			return fmt.Errorf("invalid attestation status %q", value)
		}

	case FieldNameField,
		FieldShortName,
		FieldOrganizationID,
		FieldOwnerID,
		FieldPurpose:
		_, err := v.StringValue()
		return err

	case FieldProtectionRequired,
		FieldCyberCenterRequired:
		_, err := v.BoolValue()
		return err

	default:
		return fmt.Errorf("field %q cannot be changed", field)
	}
}
