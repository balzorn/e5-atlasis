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

func (a InformationAsset) FieldValue(field FieldName) (any, error) {
	switch field {
	case FieldType:
		return a.Type, nil
	case FieldNameField:
		return a.Name, nil
	case FieldShortName:
		return a.ShortName, nil
	case FieldStatus:
		return a.Status, nil
	case FieldOrganizationID:
		return a.OrganizationID, nil
	case FieldOwnerID:
		return a.OwnerID, nil
	case FieldPurpose:
		return a.Purpose, nil
	case FieldCriticality:
		return a.Criticality, nil
	case FieldRiskLevel:
		return a.RiskLevel, nil
	case FieldProtectionRequired:
		return a.Security.ProtectionRequired, nil
	case FieldProtectionStatus:
		return a.Security.ProtectionStatus, nil
	case FieldAttestationStatus:
		return a.Security.AttestationStatus, nil
	case FieldCyberCenterRequired:
		return a.Security.CyberCenterRequired, nil
	default:
		return nil, fmt.Errorf("unknown field %q", field)
	}
}
