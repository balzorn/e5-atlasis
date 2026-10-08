package asset

import "fmt"

func (a *InformationAsset) ApplyField(
	field FieldName,
	value FieldValue,
) error {
	if err := value.ValidateFor(field); err != nil {
		return fmt.Errorf("invalid value for field %q: %w", field, err)
	}

	switch field {
	case FieldType:
		v, _ := value.StringValue()
		a.Type = AssetType(v)

	case FieldNameField:
		v, _ := value.StringValue()
		a.Name = v

	case FieldShortName:
		v, _ := value.StringValue()
		a.ShortName = v

	case FieldStatus:
		v, _ := value.StringValue()
		target := AssetStatus(v)

		if !a.CanTransitionTo(target) && a.Status != target {
			return fmt.Errorf(
				"invalid asset status transition: %s -> %s",
				a.Status,
				target,
			)
		}

		a.Status = target

	case FieldOrganizationID:
		v, _ := value.StringValue()
		a.OrganizationID = v

	case FieldOwnerID:
		v, _ := value.StringValue()
		a.OwnerID = v

	case FieldPurpose:
		v, _ := value.StringValue()
		a.Purpose = v

	case FieldCriticality:
		v, _ := value.StringValue()
		a.Criticality = Criticality(v)

	case FieldRiskLevel:
		v, _ := value.StringValue()
		a.RiskLevel = RiskLevel(v)

	case FieldProtectionRequired:
		v, _ := value.BoolValue()
		a.Security.ProtectionRequired = v

	case FieldProtectionStatus:
		v, _ := value.StringValue()
		a.Security.ProtectionStatus = ProtectionStatus(v)

	case FieldAttestationStatus:
		v, _ := value.StringValue()
		a.Security.AttestationStatus = AttestationStatus(v)

	case FieldCyberCenterRequired:
		v, _ := value.BoolValue()
		a.Security.CyberCenterRequired = v

	default:
		return fmt.Errorf("field %q cannot be changed", field)
	}

	return a.Validate()
}
