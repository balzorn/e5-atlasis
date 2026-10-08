package asset

import "fmt"

type InformationAsset struct {
	ID             AssetID
	Type           AssetType
	Name           string
	ShortName      string
	Status         AssetStatus
	OrganizationID string
	OwnerID        string
	Purpose        string
	Criticality    Criticality
	RiskLevel      RiskLevel
	Security       SecurityProfile
	CurrentVersion Version
}

func (a InformationAsset) Validate() error {
	if _, err := ParseAssetID(a.ID.String()); err != nil {
		return err
	}

	return a.ValidateAttributes()
}

func (a InformationAsset) ValidateAttributes() error {
	switch a.Type {
	case AssetTypeInformationSystem,
		AssetTypeInformationInfrastructureObject:
	default:
		return fmt.Errorf("invalid asset type %q", a.Type)
	}

	if a.Name == "" {
		return fmt.Errorf("asset name is required")
	}

	if a.OrganizationID == "" {
		return fmt.Errorf("organization ID is required")
	}

	if a.OwnerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	switch a.Status {
	case AssetStatusDraft, AssetStatusActive, AssetStatusSuspended, AssetStatusRetired:
	default:
		return fmt.Errorf("invalid asset status %q", a.Status)
	}

	switch a.Criticality {
	case CriticalityLow, CriticalityMedium, CriticalityHigh, CriticalityCritical:
	default:
		return fmt.Errorf("invalid asset criticality %q", a.Criticality)
	}

	switch a.RiskLevel {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical:
	default:
		return fmt.Errorf("invalid asset risk level %q", a.RiskLevel)
	}

	switch a.Security.ProtectionStatus {
	case ProtectionStatusNotRequired, ProtectionStatusRequired, ProtectionStatusInProgress, ProtectionStatusImplemented:
	default:
		return fmt.Errorf("invalid protection status %q", a.Security.ProtectionStatus)
	}

	switch a.Security.AttestationStatus {
	case AttestationStatusNotRequired, AttestationStatusRequired, AttestationStatusInProgress, AttestationStatusAttested, AttestationStatusExpired:
	default:
		return fmt.Errorf("invalid attestation status %q", a.Security.AttestationStatus)
	}

	return nil
}

func (a InformationAsset) CanTransitionTo(target AssetStatus) bool {
	switch a.Status {
	case AssetStatusDraft:
		return target == AssetStatusActive ||
			target == AssetStatusRetired

	case AssetStatusActive:
		return target == AssetStatusSuspended ||
			target == AssetStatusRetired

	case AssetStatusSuspended:
		return target == AssetStatusActive ||
			target == AssetStatusRetired

	case AssetStatusRetired:
		return false

	default:
		return false
	}
}
