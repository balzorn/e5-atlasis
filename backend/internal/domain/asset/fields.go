package asset

import "fmt"

type FieldName string

const (
	FieldType                FieldName = "type"
	FieldNameField           FieldName = "name"
	FieldShortName           FieldName = "short_name"
	FieldStatus              FieldName = "status"
	FieldOrganizationID      FieldName = "organization_id"
	FieldOwnerID             FieldName = "owner_id"
	FieldPurpose             FieldName = "purpose"
	FieldCriticality         FieldName = "criticality"
	FieldRiskLevel           FieldName = "risk_level"
	FieldProtectionRequired  FieldName = "protection_required"
	FieldProtectionStatus    FieldName = "protection_status"
	FieldAttestationStatus   FieldName = "attestation_status"
	FieldCyberCenterRequired FieldName = "cyber_center_required"
)

func ParseFieldName(value string) (FieldName, error) {
	field := FieldName(value)

	switch field {
	case FieldType,
		FieldNameField,
		FieldShortName,
		FieldStatus,
		FieldOrganizationID,
		FieldOwnerID,
		FieldPurpose,
		FieldCriticality,
		FieldRiskLevel,
		FieldProtectionRequired,
		FieldProtectionStatus,
		FieldAttestationStatus,
		FieldCyberCenterRequired:
		return field, nil

	default:
		return "", fmt.Errorf("field %q cannot be changed", value)
	}
}

func (a InformationAsset) FieldValue(field FieldName) (FieldValue, error) {
	switch field {
	case FieldType:
		return NewStringFieldValue(string(a.Type)), nil
	case FieldNameField:
		return NewStringFieldValue(a.Name), nil
	case FieldShortName:
		return NewStringFieldValue(a.ShortName), nil
	case FieldStatus:
		return NewStringFieldValue(string(a.Status)), nil
	case FieldOrganizationID:
		return NewStringFieldValue(a.OrganizationID), nil
	case FieldOwnerID:
		return NewStringFieldValue(a.OwnerID), nil
	case FieldPurpose:
		return NewStringFieldValue(a.Purpose), nil
	case FieldCriticality:
		return NewStringFieldValue(string(a.Criticality)), nil
	case FieldRiskLevel:
		return NewStringFieldValue(string(a.RiskLevel)), nil
	case FieldProtectionRequired:
		return NewBoolFieldValue(a.Security.ProtectionRequired), nil
	case FieldProtectionStatus:
		return NewStringFieldValue(string(a.Security.ProtectionStatus)), nil
	case FieldAttestationStatus:
		return NewStringFieldValue(string(a.Security.AttestationStatus)), nil
	case FieldCyberCenterRequired:
		return NewBoolFieldValue(a.Security.CyberCenterRequired), nil
	default:
		return FieldValue{}, fmt.Errorf("unknown field %q", field)
	}
}
