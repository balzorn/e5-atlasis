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
}

func (a InformationAsset) Validate() error {
	if _, err := ParseAssetID(a.ID.String()); err != nil {
		return err
	}

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
